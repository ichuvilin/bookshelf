package main

import (
	"bookshelf/worker-service/internal/config"
	"bookshelf/worker-service/internal/queue"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("worker-service starting...")

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
