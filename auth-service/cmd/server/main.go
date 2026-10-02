package main

import (
	"bookshelf/auth-service/internal/config"
	"bookshelf/auth-service/internal/repository"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

	_ = repository.NewUserRepository(db)

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"service":  "auth-service",
			"database": "connected",
		}); err != nil {
			return
		}

	})

	err = http.ListenAndServe(cfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
