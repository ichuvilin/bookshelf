package handler

import (
	"bookshelf/books-service/internal/client"
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type HealthResponse struct {
	Status    string           `json:"status"` // "ok", "unhealthy"
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
	Timestamp string           `json:"timestamp"`
}

type Check struct {
	Status   string `json:"status"`   // "ok", "error"
	Duration string `json:"duration"` // "2ms"
	Error    string `json:"error,omitempty"`
}

type ReadyResponse struct {
	Ready     bool             `json:"ready"`
	Service   string           `json:"service"`
	Checks    map[string]Check `json:"checks"`
	Timestamp string           `json:"timestamp"`
}

type HealthHandler struct {
	db         *sqlx.DB
	version    string
	authClient *client.AuthClient
	ioClient   *client.MinIOClient
}

func NewHealthHandler(db *sqlx.DB, version string, authClient *client.AuthClient, ioClient *client.MinIOClient) *HealthHandler {
	return &HealthHandler{
		db:         db,
		version:    version,
		authClient: authClient,
		ioClient:   ioClient,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := map[string]Check{
		"database": h.checkDatabase(),
		"io":       h.checkIO(),
	}

	status := "ok"
	httpStatus := http.StatusOK

	for _, check := range checks {
		if check.Status == "error" {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Service:   "books-service",
		Version:   h.version,
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, httpStatus, response)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]Check)

	allReady := true

	checks["database"] = h.checkDatabase()
	if checks["database"].Status != "ok" {
		allReady = false
	}

	checks["auth-service"] = h.checkAuthService()
	if checks["auth-service"].Status != "ok" {
		allReady = false
	}

	checks["io"] = h.checkIO()
	if checks["io"].Status != "ok" {
		allReady = false
	}

	status := http.StatusOK
	if !allReady {
		status = http.StatusServiceUnavailable
	}

	response := ReadyResponse{
		Ready:     allReady,
		Service:   "books-service",
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, status, response)
}

func (h *HealthHandler) checkAuthService() Check {
	start := time.Now()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	resp, err := h.authClient.Health(ctx)

	duration := time.Since(start).String()

	if err != nil {
		return Check{
			Status:   "error",
			Duration: duration,
			Error:    err.Error(),
		}
	}

	if resp.Status != "ok" {
		return Check{
			Status:   "error",
			Duration: duration,
			Error:    "auth-service is unhealthy",
		}
	}

	return Check{
		Status:   "ok",
		Duration: duration,
	}
}

func (h *HealthHandler) checkDatabase() Check {
	start := time.Now()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	err := h.db.PingContext(ctx)

	duration := time.Since(start).String()

	if err != nil {
		return Check{
			Status:   "error",
			Duration: duration,
			Error:    err.Error(),
		}
	}

	return Check{
		Status:   "ok",
		Duration: duration,
	}
}

func (h *HealthHandler) checkIO() Check {
	start := time.Now()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	err := h.ioClient.HealthCheck(ctx)

	duration := time.Since(start).String()

	if err != nil {
		return Check{
			Status:   "error",
			Duration: duration,
			Error:    err.Error(),
		}
	}

	return Check{
		Status:   "ok",
		Duration: duration,
	}
}
