package main

import (
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/handler"
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/repository"
	"bookshelf/worker-service/internal/storage"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("worker-service starting...")

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	log.Println("Connected to database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	coverRepo := repository.NewCoverRepository(db)

	storage, err := storage.NewMinIOStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinioBucket, cfg.MinIOPublicEndpoint, cfg.MinIOUseSSL)
	log.Println("Connected to storage")
	if err != nil {
		log.Fatal(err)
	}

	imageHandler := handler.NewImageHandler(storage, coverRepo)

	consumer, err := queue.NewConsumer(cfg.RabbitMQURL)
	log.Println("Connected to rabbit")
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	consumer.RegisterHandler("image_compress", imageHandler.HandleImageCompress)

	log.Println("Starting consumer...")
	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}

	healthHandler := handler.NewHealthHandler(db, "1.0.0", storage, consumer)

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

	log.Println("Start web server")
	err = http.ListenAndServe(cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
