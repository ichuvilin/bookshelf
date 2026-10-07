package main

import (
	"bookshelf/worker-service/internal/config"
	"log"
)

func main() {
	_ = config.LoadConfig()
	log.Println("worker-service starting...")
}
