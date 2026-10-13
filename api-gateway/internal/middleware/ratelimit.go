package middleware

import (
	"bookshelf/api-gateway/internal/domain"
	"bookshelf/api-gateway/internal/ratelimit"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RateLimitConfig struct {
	Limiter         *ratelimit.RateLimiter
	AnonymousLimit  int
	AuthorizedLimit int
	UploadLimit     int
}

func RateLimitMiddleware(cfg *RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}

			if cfg == nil || cfg.Limiter == nil {
				http.Error(w, "rate limiter is not configured", http.StatusInternalServerError)
				return
			}

			key, limit := getRateLimitKeyAndLimit(r, cfg)
			result, err := cfg.Limiter.Allow(r.Context(), key, limit)
			if err != nil {
				slog.Error("rate limit check failed",
					"error", err,
					"rate_limit_key", key,
				)
				http.Error(
					w,
					"rate limit service unavailable",
					http.StatusServiceUnavailable,
				)
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

func getRateLimitKeyAndLimit(
	r *http.Request,
	cfg *RateLimitConfig,
) (string, int) {
	ip := getClientIP(r)

	// Загрузка обложки имеет приоритет над обычным лимитом.
	if r.Method == http.MethodPost &&
		strings.Contains(r.URL.Path, "/cover") {
		if token := bearerToken(r); token != "" {
			return "upload:auth:" + hashToken(token), cfg.UploadLimit
		}
		return "upload:ip:" + ip, cfg.UploadLimit
	}

	// Авторизованный клиент.
	if token := bearerToken(r); token != "" {
		return "auth:user:" + hashToken(token), cfg.AuthorizedLimit
	}

	// Анонимный клиент.
	return "anon:ip:" + ip, cfg.AnonymousLimit
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "

	if len(header) <= len(prefix) ||
		!strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return ""
	}

	return token
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:12])
}

func getClientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip := strings.TrimSpace(strings.SplitN(forwardedFor, ",", 2)[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(realIP) != nil {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
