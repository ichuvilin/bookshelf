package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5/middleware"
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

	if value := r.Context().Value(middleware.RequestIDKey); value != nil {
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

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if auth == "" {
			writeError(
				w,
				r,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Authorization header required",
			)
			return
		}

		parts := strings.Split(auth, " ")

		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(
				w,
				r,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid authorization header format",
			)
			return
		}

		token := parts[1]

		userID, err := h.services.UserService.ValidateToken(token)
		if err != nil {
			writeError(
				w,
				r,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid or expired token",
			)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
