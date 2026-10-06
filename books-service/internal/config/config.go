package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	AuthServiceURL string
	ServiceKey     string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:postgres@localhost:5433/books?sslmode=disable"
	}

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8081"
	}

	svcKey := os.Getenv("SERVICE_KEY")
	if svcKey == "" {
		svcKey = "dev-service-key"
	}

	return &Config{
		Port:           fmt.Sprintf(":%s", port),
		DatabaseURL:    dbUrl,
		AuthServiceURL: authServiceURL,
		ServiceKey:     svcKey,
	}
}
