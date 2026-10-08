package main

import (
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/storage"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("worker-service starting...")

	_, err := storage.NewMinIOStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinioBucket, cfg.MinIOPublicEndpoint, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := queue.NewConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	consumer.RegisterHandler("image_compress", func(body []byte) error {
		log.Printf("Received message: %s", string(body))
		return nil
	})

	log.Println("Starting consumer...")
	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
