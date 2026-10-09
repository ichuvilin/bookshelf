package main

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/handler"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Connected to database")

	rabbit, err := client.NewRabbitMQClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	authClient := client.NewAuthClient(cfg.AuthServiceURL, 10*time.Second, cfg.ServiceKey, cfg.AuthServiceRetries, cfg.AuthServiceTimeout)

	ioClient, err := client.NewMinIOClient(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinioBucket, cfg.MinIOPublicEndpoint, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	coverRepo := repository.NewCoverRepository(db)

	bookSvc := service.NewBookService(bookRepo)
	reviewSvc := service.NewReviewService(reviewRepo, bookRepo, authClient)
	coverSvc := service.NewCoverService(coverRepo, bookRepo, ioClient, rabbit)

	bookHandler := handler.NewBookHandler(bookSvc)
	reviewHandler := handler.NewReviewHandler(reviewSvc)
	healthHandler := handler.NewHealthHandler(db, "1.0.0", authClient, ioClient)
	coverHandler := handler.NewCoverHandler(coverSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5174"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/books", bookHandler.List)
		r.Get("/books/{id}", bookHandler.GetByID)
		r.Get("/books/{book_id}/reviews", reviewHandler.List)
		r.Get("/books/{id}/cover", coverHandler.GetBookCover)
		r.Get("/books/{id}/cover/status", coverHandler.GetBookCoverStatus)

		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(authClient))

			r.Post("/books", bookHandler.Create)
			r.Put("/books/{id}", bookHandler.Update)
			r.Delete("/books/{id}", bookHandler.Delete)

			r.Post("/books/{book_id}/reviews", reviewHandler.Create)
			r.Put("/reviews/{id}", reviewHandler.Update)
			r.Delete("/reviews/{id}", reviewHandler.Delete)

			r.Post("/books/{id}/cover", coverHandler.UploadBookCover)
			r.Delete("/books/{id}/cover", coverHandler.DeleteBookCover)
		})
	})

	err = http.ListenAndServe(cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
