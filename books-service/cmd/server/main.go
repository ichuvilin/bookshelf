package main

import (
	"bookshelf/books-service/internal/config"
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "books-service",
		}); err != nil {
			return
		}
	})

	err := http.ListenAndServe(cfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
