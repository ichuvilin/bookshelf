package main

import (
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	db.Close()

	log.Println("Connected to database")

	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	_ = service.NewBookService(bookRepo)
	_ = service.NewReviewService(reviewRepo, bookRepo)

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "books-service",
		}); err != nil {
			return
		}
	})

	err = http.ListenAndServe(cfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
