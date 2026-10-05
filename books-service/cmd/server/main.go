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

	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	bookSvc := service.NewBookService(bookRepo)
	reviewSvc := service.NewReviewService(reviewRepo, bookRepo)

	bookHandler := handler.NewBookHandler(bookSvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc)

	_ = client.NewHTTPClient(cfg.AuthServiceURL, 5*time.Second)

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

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/books", bookHandler.List)
		r.Post("/books", bookHandler.Create)
		r.Get("/books/{id}", bookHandler.GetByID)
		r.Put("/books/{id}", bookHandler.Update)
		r.Delete("/books/{id}", bookHandler.Delete)

		r.Post("/books/{book_id}/reviews", reviewHandler.Create)
		r.Get("/books/{book_id}/reviews", reviewHandler.List)
		r.Put("/reviews/{id}", reviewHandler.Update)
		r.Delete("/reviews/{id}", reviewHandler.Delete)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "books-service",
		}); err != nil {
			return
		}
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
