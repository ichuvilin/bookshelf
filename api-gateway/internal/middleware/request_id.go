package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")

		if reqID == "" {
			reqID = uuid.New().String()
		}

		r.Header.Set("X-Request-Id", reqID)
		w.Header().Set("X-Request-Id", reqID)

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
