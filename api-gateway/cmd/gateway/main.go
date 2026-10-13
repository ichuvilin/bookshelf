package main

import (
	"bookshelf/api-gateway/internal/config"
	"bookshelf/api-gateway/internal/handler"
	m "bookshelf/api-gateway/internal/middleware"
	"bookshelf/api-gateway/internal/proxy"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	serviceProxy := proxy.New(cfg.AuthServiceURL, cfg.BooksServiceURL)

	authHandler := handler.NewAuthHandler(serviceProxy)
	booksHandler := handler.NewBooksHandler(serviceProxy)
	reviewsHandler := handler.NewReviewsHandler(serviceProxy)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(m.RequestID)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5174"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		}); err != nil {
			return
		}
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/logout", authHandler.Logout)

		// Users
		r.Get("/users/me", authHandler.GetCurrentUser)
		r.Get("/users/{userId}", authHandler.GetUser)

		// Books
		r.Get("/books", booksHandler.ListBooks)
		r.Post("/books", booksHandler.CreateBook)
		r.Get("/books/{bookId}", booksHandler.GetBook)
		r.Put("/books/{bookId}", booksHandler.UpdateBook)
		r.Patch("/books/{bookId}", booksHandler.UpdateBook)
		r.Delete("/books/{bookId}", booksHandler.DeleteBook)

		// Reviews
		r.Get("/books/{bookId}/reviews", reviewsHandler.ListReviews)
		r.Post("/books/{bookId}/reviews", reviewsHandler.CreateReview)
		r.Get("/books/{bookId}/reviews/{reviewId}", reviewsHandler.GetReview)
		r.Put("/books/{bookId}/reviews/{reviewId}", reviewsHandler.UpdateReview)
		r.Patch("/books/{bookId}/reviews/{reviewId}", reviewsHandler.UpdateReview)
		r.Delete("/books/{bookId}/reviews/{reviewId}", reviewsHandler.DeleteReview)
	})

	log.Println("API Gateway listening on :8000")
	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatal("Failed started api gateway", err)
	}
}
