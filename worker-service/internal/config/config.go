package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQURL         string // URL для подключения к RabbitMQ
	MinIOEndpoint       string // URL для подключения к MinIO
	MinIOAccessKey      string // Access key для MinIO
	MinIOSecretKey      string // Secret key для MinIO
	MinioBucket         string // Имя bucket в MinIO
	MinIOPublicEndpoint string // http://localhost:9000
	MinIOUseSSL         bool   // false для локальной разработки
	DatabaseURL         string // URL для подключения к PostgreSQL (books_db)
}

func LoadConfig() *Config {
	mqurl := os.Getenv("RABBIT_MQURL")
	if mqurl == "" {
		mqurl = "amqp://guest:guest@localhost:5672/"
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/books?sslmode=disable"
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
		RabbitMQURL:         mqurl,
		DatabaseURL:         dbURL,
		MinIOEndpoint:       minioEndpoint,
		MinIOAccessKey:      minioAccessKey,
		MinIOSecretKey:      minioSecretKey,
		MinioBucket:         minioBucket,
		MinIOPublicEndpoint: minioPublicEndpoint,
		MinIOUseSSL:         minioUseSSL,
	}
}
