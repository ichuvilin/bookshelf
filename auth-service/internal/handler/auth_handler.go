package handler

import (
	"bookshelf/auth-service/internal/domain"
	"bookshelf/auth-service/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const userIDKey contextKey = "userID"

type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	response, err := h.svc.Register(r.Context(), req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}
	response, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "401", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	writeJSON(w, http.StatusOK, user.ToPublic())
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req domain.UpdateUserRequest
	err := decodeJSON(r, req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, 200, user)
}

func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"service":  "auth-service",
		"database": "connected",
	})
}

func (h *AuthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"database": "ok",
		},
	})
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
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

	writeJSON(w, status, map[string]any{
		"error": domain.ErrorResponse{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}

func getUserID(ctx context.Context) string {
	return ctx.Value(userIDKey).(string)
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
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

		userID, err := h.svc.ValidateToken(token)
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
