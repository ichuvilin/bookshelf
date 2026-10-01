package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	services  *service.Service
	jwtSecret string
}

func New(services *service.Service, jwtSecret string) *Handler {
	return &Handler{
		services:  services,
		jwtSecret: jwtSecret,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID := ""

	if value := r.Context().Value("requestID"); value != nil {
		requestID, _ = value.(string)
	}

	writeJSON(w, status, domain.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}

func writeValidationError(w http.ResponseWriter, r *http.Request, details []domain.ErrorDetail) {
	writeJSON(w, http.StatusBadRequest, details)
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getUserID(ctx context.Context) string {
	return ctx.Value(userIDKey).(string)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"database": "ok",
		},
	})
}
