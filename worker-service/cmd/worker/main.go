package main

import (
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/handler"
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/repository"
	"bookshelf/worker-service/internal/storage"
	"log"

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

	select {}
}
