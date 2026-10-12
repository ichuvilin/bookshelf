package main

import (
	"bookshelf/api-gateway/internal/config"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		}); err != nil {
			return
		}
	})

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatal("Failed started api gateway", err)
	}
}
