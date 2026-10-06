package handler

import (
	"bookshelf/books-service/internal/domain"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

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
