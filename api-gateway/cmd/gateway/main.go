package main

import (
	"bookshelf/api-gateway/internal/cache"
	"bookshelf/api-gateway/internal/config"
	"bookshelf/api-gateway/internal/handler"
	m "bookshelf/api-gateway/internal/middleware"
	"bookshelf/api-gateway/internal/proxy"
	"bookshelf/api-gateway/internal/ratelimit"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	logger.Info("Config loaded")

	c, err := cache.New(cfg.RedisURL)
	if err != nil {
		logger.Error("error during connection to redis", err)
		os.Exit(0)
	}
	defer c.Close()

	rl := ratelimit.New(
		c.GetClient(),
		time.Duration(cfg.RateLimitWindow)*time.Second,
	)

	rlConfig := &m.RateLimitConfig{
		Limiter: rl,
		Limit:   100,
	}

	serviceProxy := proxy.New(cfg.AuthServiceURL, cfg.BooksServiceURL)

	authHandler := handler.NewAuthHandler(serviceProxy)
	booksHandler := handler.NewBooksHandler(serviceProxy)
	reviewsHandler := handler.NewReviewsHandler(serviceProxy)

	r := chi.NewRouter()
	r.Use(m.CORS)
	r.Use(middleware.Recoverer)
	r.Use(m.RequestID)
	r.Use(m.Logging)
	r.Use(m.RateLimitMiddleware(rlConfig))

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

	logger.Info("API Gateway listening on :8000")
	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		logger.Error("Failed started api gateway", err)
	}
}
