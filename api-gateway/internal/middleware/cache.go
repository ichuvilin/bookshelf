package middleware

import (
	"log/slog"
	"net/http"
	"time"

	cachepkg "bookshelf/api-gateway/internal/cache"
)

type CacheConfig struct {
	Cache     *cachepkg.Cache
	TTL       time.Duration
	KeyPrefix string
}

func CacheMiddleware(cfg *CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Кэшируем только GET.
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			if cfg == nil || cfg.Cache == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := cachepkg.GenerateKey(
				cfg.KeyPrefix,
				r.URL.Path,
				r.URL.RawQuery,
			)

			data, found := cfg.Cache.Get(r.Context(), key)
			if found {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")

				if _, err := w.Write(data); err != nil {
					slog.Error("failed to write cached response",
						"error", err,
						"cache_key", key,
					)
				}
				return
			}

			w.Header().Set("X-Cache", "MISS")
			next.ServeHTTP(w, r)
		})
	}
}
