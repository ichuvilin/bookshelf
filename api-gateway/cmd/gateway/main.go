package main

import (
	"bookshelf/api-gateway/internal/config"
	"bookshelf/api-gateway/internal/proxy"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	serviceProxy := proxy.New(cfg.AuthServiceURL, cfg.BooksServiceURL)

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

	r.Route("/api/v1", func(r chi.Router) {
		r.Handle("/auth/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceProxy.ProxyToAuth(w, r, r.URL.Path)
		}))
		r.Handle("/users/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceProxy.ProxyToAuth(w, r, r.URL.Path)
		}))
		r.Handle("/books/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceProxy.ProxyToBooks(w, r, r.URL.Path)
		}))
	})

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatal("Failed started api gateway", err)
	}
}
