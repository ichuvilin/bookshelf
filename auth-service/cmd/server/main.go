package main

import (
	"bookshelf/auth-service/internal/config"
	"log"
	"net/http"
)

func main() {
	load := config.Load()

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"ok\"}"))
	})

	err := http.ListenAndServe(load.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
