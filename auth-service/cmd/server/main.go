package main

import (
	"bookshelf/auth-service/internal/config"
	"bookshelf/auth-service/internal/handler"
	"bookshelf/auth-service/internal/repository"
	"bookshelf/auth-service/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("error during connect to database", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Second)

	log.Println("Connected to database")

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, cfg.JWTSecret)
	h := handler.NewAuthHandler(svc)
	internalHandler := handler.NewInternalHandler(svc)
	healthHandler := handler.NewHealthHandler(db, "1.0.0")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5174"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			r.Get("/users/me", h.GetMe)
			r.Put("/users/me", h.UpdateMe)
		})
	})

	r.Route("/internal/v1", func(r chi.Router) {
		r.Use(handler.ServiceKeyMiddleware(cfg.ServiceKey))

		r.Post("/auth/verify", internalHandler.VerifyToken)
		r.Post("/users/batch", internalHandler.GetUsersByIDs)
	})

	err = http.ListenAndServe(cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
