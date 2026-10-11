package main

import (
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/handler"
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/repository"
	"bookshelf/worker-service/internal/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("worker-service starting...")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	log.Println("Connected to database")
	defer db.Close()

	coverRepo := repository.NewCoverRepository(db)

	storageClient, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinioBucket,
		cfg.MinIOPublicEndpoint,
		cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("connect storage: %v", err)
	}
	log.Println("Connected to storage")

	imageHandler := handler.NewImageHandler(storageClient, coverRepo)

	consumer, err := queue.NewConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("connect RabbitMQ: %v", err)
	}

	consumer.RegisterHandler("image_compress", imageHandler.HandleImageCompress)

	if err := consumer.Start(); err != nil {
		log.Fatalf("start consumer: %v", err)
	}
	log.Println("Consumer started")

	healthHandler := handler.NewHealthHandler(db, "1.0.0", storageClient, consumer)

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

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Println("Starting web server")
		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
	}

	// Сначала прекращаем consumer и ждём завершения обработки.
	log.Println("Stopping consumer...")
	if err := consumer.Stop(); err != nil {
		log.Printf("Stop consumer: %v", err)
	}

	// Затем останавливаем HTTP-сервер.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("HTTP close error: %v", closeErr)
		}
	}

	log.Println("Worker-service stopped")
}
