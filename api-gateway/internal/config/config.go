package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	AuthServiceURL  string
	BooksServiceURL string
	RedisURL        string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	authSvcUrl := os.Getenv("AUTH_SERVICE_URL")
	if authSvcUrl == "" {
		authSvcUrl = "http://localhost:8081"
	}

	bookSvcUrl := os.Getenv("BOOKS_SERVICE_URL")
	if bookSvcUrl == "" {
		bookSvcUrl = "http://localhost:8082"
	}

	redisURL := os.Getenv("REDIS_URL")

	return &Config{
		Port:            fmt.Sprintf(":%s", port),
		AuthServiceURL:  authSvcUrl,
		BooksServiceURL: bookSvcUrl,
		RedisURL:        redisURL,
	}
}
