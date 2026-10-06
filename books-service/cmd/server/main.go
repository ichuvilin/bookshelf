package main

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/handler"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"encoding/json"
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
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Connected to database")

	authClient := client.NewAuthClient(cfg.AuthServiceURL, 10*time.Second, cfg.ServiceKey, cfg.AuthServiceRetries, cfg.AuthServiceTimeout)

	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	bookSvc := service.NewBookService(bookRepo)
	reviewSvc := service.NewReviewService(reviewRepo, bookRepo, authClient)

	bookHandler := handler.NewBookHandler(bookSvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc)
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/books", bookHandler.List)
		r.Get("/books/{id}", bookHandler.GetByID)
		r.Get("/books/{book_id}/reviews", reviewHandler.List)

		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(authClient))

			r.Post("/books", bookHandler.Create)
			r.Put("/books/{id}", bookHandler.Update)
			r.Delete("/books/{id}", bookHandler.Delete)

			r.Post("/books/{book_id}/reviews", reviewHandler.Create)
			r.Put("/reviews/{id}", reviewHandler.Update)
			r.Delete("/reviews/{id}", reviewHandler.Delete)
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "books-service",
		}); err != nil {
			return
		}
	})

	err = http.ListenAndServe(cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
