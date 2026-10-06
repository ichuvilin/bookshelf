package handler

import (
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

type HealthHandler struct {
	db      *sqlx.DB
	version string
}

func NewHealthHandler(db *sqlx.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		version: version,
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
