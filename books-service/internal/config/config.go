package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	AuthServiceURL      string
	ServiceKey          string
	AuthServiceTimeout  time.Duration
	AuthServiceRetries  int
	MinIOEndpoint       string // localhost:9000
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinioBucket         string // bookshelf-covers
	MinIOPublicEndpoint string // http://localhost:9000
	MinIOUseSSL         bool   // false для локальной разработки
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

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}

	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}

	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "bookshelf-covers"
	}

	minioPublicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if minioPublicEndpoint == "" {
		minioPublicEndpoint = "http://localhost:9000"
	}

	minioUseSSL := false
	if value := os.Getenv("MINIO_USE_SSL"); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			minioUseSSL = parsed
		}
	}

	return &Config{
		Port:                fmt.Sprintf(":%s", port),
		DatabaseURL:         dbUrl,
		AuthServiceURL:      authServiceURL,
		ServiceKey:          svcKey,
		AuthServiceTimeout:  authServiceTimeout,
		AuthServiceRetries:  authServiceRetries,
		MinIOEndpoint:       minioEndpoint,
		MinIOAccessKey:      minioAccessKey,
		MinIOSecretKey:      minioSecretKey,
		MinioBucket:         minioBucket,
		MinIOPublicEndpoint: minioPublicEndpoint,
		MinIOUseSSL:         minioUseSSL,
	}
}
