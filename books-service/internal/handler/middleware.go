package handler

import (
	"bookshelf/books-service/internal/client"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "user_id"

func AuthMiddleware(authClient *client.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				WriteError(
					w,
					r,
					http.StatusUnauthorized,
					"401",
					"missing authorization header",
				)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
				WriteError(
					w,
					r,
					http.StatusUnauthorized,
					"401",
					"invalid authorization header",
				)
				return
			}

			token := parts[1]

			resp, err := authClient.VerifyToken(r.Context(), token)
			if err != nil {
				WriteError(
					w,
					r,
					http.StatusServiceUnavailable,
					"503",
					"auth service unavailable",
				)
				return
			}

			if !resp.Valid {
				WriteError(
					w,
					r,
					http.StatusUnauthorized,
					"401",
					"invalid token",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				resp.UserID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getUserID(ctx context.Context) string {
	return ctx.Value(userIDKey).(string)
}
