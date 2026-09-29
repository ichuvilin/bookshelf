package main

import (
	"net/http"

	"github.com/bookshelf/monolith/internal/config"
)

func main() {
	cfg := config.Load()

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"ok\"}"))
	})

	http.ListenAndServe(cfg.Port, nil)
}
