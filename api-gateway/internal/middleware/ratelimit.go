package middleware

import (
	"bookshelf/api-gateway/internal/domain"
	"bookshelf/api-gateway/internal/ratelimit"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type RateLimitConfig struct {
	Limiter *ratelimit.RateLimiter
	Limit   int
}

func RateLimitMiddleware(cfg *RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg == nil || cfg.Limiter == nil || cfg.Limit <= 0 {
				http.Error(w, "rate limiter is not configured", http.StatusInternalServerError)
				return
			}

			clientIP := getRateLimitClientIP(r)
			result, err := cfg.Limiter.Allow(r.Context(), clientIP, cfg.Limit)
			if err != nil {
				slog.Error("rate limit check failed",
					"error", err,
					"client_ip", clientIP,
				)
				http.Error(w, "rate limit service unavailable", http.StatusServiceUnavailable)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

			if !result.Allowed {
				retryAfter := time.Until(result.ResetAt).Seconds()
				retryAfterSeconds := int64(retryAfter)
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}

				w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
				w.Header().Set("Content-Type", "application/json")

				requestID, _ := r.Context().Value(RequestIDKey).(string)

				w.WriteHeader(http.StatusTooManyRequests)
				if err := json.NewEncoder(w).Encode(domain.ErrorResponse{
					Code:      "RATE_LIMIT_EXCEEDED",
					Message:   "Too many requests",
					RequestID: requestID,
				}); err != nil {
					slog.Error("failed to write rate limit response", "error", err)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getRateLimitClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
