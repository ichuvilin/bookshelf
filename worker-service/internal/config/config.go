package config

import "os"

type Config struct {
	RabbitMQURL string // URL для подключения к RabbitMQ
	MinIOURL    string // URL для подключения к MinIO
	MinIOAccess string // Access key для MinIO
	MinIOSecret string // Secret key для MinIO
	MinioBucket string // Имя bucket в MinIO
	DatabaseURL string // URL для подключения к PostgreSQL (books_db)
}

func LoadConfig() *Config {
	mqurl := os.Getenv("RABBIT_MQURL")
	if mqurl == "" {
		mqurl = "amqp://guest:guest@localhost:5672/"
	}
	minIOURL := os.Getenv("MINIO_URL")
	if minIOURL == "" {
		minIOURL = "url"
	}
	minIOAccess := os.Getenv("MINIO_ACCESS")
	if minIOAccess == "" {
		minIOAccess = "access"
	}
	minIOSecret := os.Getenv("MINIO_SECRET")
	if minIOSecret == "" {
		minIOSecret = "secret"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "image"
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/books?sslmode=disable"
	}

	return &Config{
		RabbitMQURL: mqurl,
		MinIOURL:    minIOURL,
		MinIOAccess: minIOAccess,
		MinIOSecret: minIOSecret,
		MinioBucket: bucket,
		DatabaseURL: dbURL,
	}
}
