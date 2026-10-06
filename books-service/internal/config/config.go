package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	AuthServiceURL     string
	ServiceKey         string
	AuthServiceTimeout time.Duration
	AuthServiceRetries int
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

	authServiceTimeout := 5 * time.Second
	if value := os.Getenv("AUTH_SERVICE_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			authServiceTimeout = parsed
		}
	}

	authServiceRetries := 3
	if value := os.Getenv("AUTH_SERVICE_RETRIES"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			authServiceRetries = parsed
		}
	}

	return &Config{
		Port:               fmt.Sprintf(":%s", port),
		DatabaseURL:        dbUrl,
		AuthServiceURL:     authServiceURL,
		ServiceKey:         svcKey,
		AuthServiceTimeout: authServiceTimeout,
		AuthServiceRetries: authServiceRetries,
	}
}
