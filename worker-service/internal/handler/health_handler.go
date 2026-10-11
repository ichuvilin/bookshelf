package handler

import (
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/storage"
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db       *sqlx.DB
	version  string
	storage  *storage.MinIOStorage
	consumer *queue.Consumer
}

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
	Status    string           `json:"status"` // "ok", "error"
	Ready     bool             `json:"ready"`
	Service   string           `json:"service"`
	Checks    map[string]Check `json:"checks"`
	Timestamp string           `json:"timestamp"`
}

func NewHealthHandler(db *sqlx.DB, version string, storage *storage.MinIOStorage, consumer *queue.Consumer) *HealthHandler {
	return &HealthHandler{
		db:       db,
		version:  version,
		storage:  storage,
		consumer: consumer,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := map[string]Check{
		"database": h.checkDatabase(),
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

	checks["io"] = h.checkIO()
	if checks["io"].Status != "ok" {
		allReady = false
	}

	checks["rabbitmq"] = h.checkMq()
	if checks["rabbitmq"].Status != "ok" {
		allReady = false
	}

	status := http.StatusOK
	svcStatus := "ok"
	if !allReady {
		status = http.StatusServiceUnavailable
		svcStatus = "error"
	}

	response := ReadyResponse{
		Status:    svcStatus,
		Ready:     allReady,
		Service:   "books-service",
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	writeJSON(w, status, response)
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

	err := h.storage.HealthCheck(ctx)

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

func (h *HealthHandler) checkMq() Check {
	start := time.Now()

	err := h.consumer.HealthCheck()

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
