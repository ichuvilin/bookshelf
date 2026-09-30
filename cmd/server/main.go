package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bookshelf/monolith/internal/config"
	"github.com/bookshelf/monolith/internal/repository"
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

	_ = repository.New(db)

	log.Println("Connected to database")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"ok\"}"))
	})

	err = http.ListenAndServe(cfg.Port, r)
	if err != nil {
		fmt.Printf("error starting server: %v\n", err)
	}
}
