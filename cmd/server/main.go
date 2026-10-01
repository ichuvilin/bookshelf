package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bookshelf/monolith/internal/config"
	"github.com/bookshelf/monolith/internal/handler"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	repos := repository.New(db)
	services := service.New(repos, cfg.JWTSecret)
	handlers := handler.New(services, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", handlers.Register)
		r.Post("/auth/login", handlers.Login)
		r.Get("/books", handlers.ListBooks)
		r.Get("/books/{bookId}", handlers.GetBook)
		r.Get("/books/{bookId}/reviews", handlers.ListBookReviews)
		r.Get("/reviews/{reviewId}", handlers.GetReview)

		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)

			r.Get("/users/me", handlers.GetCurrentUser)
			r.Put("/users/me", handlers.UpdateCurrentUser)
			r.Post("/books", handlers.CreateBook)
			r.Put("/books/{bookId}", handlers.UpdateBook)
			r.Delete("/books/{bookId}", handlers.DeleteBook)
			r.Post("/books/{bookId}/reviews", handlers.CreateReview)
			r.Put("/reviews/{reviewId}", handlers.UpdateReview)
			r.Delete("/reviews/{reviewId}", handlers.DeleteReview)
		})
	})

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
