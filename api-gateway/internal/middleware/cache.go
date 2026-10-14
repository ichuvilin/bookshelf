package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	cachepkg "bookshelf/api-gateway/internal/cache"
)

type CacheConfig struct {
	Cache     *cachepkg.Cache
	TTL       time.Duration
	KeyPrefix string
}

type CachedResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	body        *bytes.Buffer
	wroteHeader bool
}

func NewCachedResponseWriter(w http.ResponseWriter) *CachedResponseWriter {
	return &CachedResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           new(bytes.Buffer),
	}
}

func (w *CachedResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *CachedResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func CacheMiddleware(cfg *CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			if cfg == nil || cfg.Cache == nil || cfg.TTL <= 0 {
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
				var cached cachepkg.CachedResponse

				if err := json.Unmarshal(data, &cached); err == nil {
					if cached.StatusCode < 200 ||
						cached.StatusCode >= 300 ||
						len(cached.Body) == 0 ||
						!json.Valid(cached.Body) {
						slog.Warn(
							"invalid cached response",
							"cache_key", key,
						)
					} else {
						for name, value := range cached.Headers {
							// Content-Length выставим по фактическому телу.
							if http.CanonicalHeaderKey(name) == "Content-Length" {
								continue
							}

							w.Header().Set(name, value)
						}

						w.Header().Set("X-Cache", "HIT")
						w.Header().Set(
							"Content-Length",
							strconv.Itoa(len(cached.Body)),
						)

						slog.Info(
							"cache response headers",
							"cache_key", key,
							"content_length", w.Header().Get("Content-Length"),
							"body_size", len(cached.Body),
							"content_type", w.Header().Get("Content-Type"),
						)

						w.WriteHeader(cached.StatusCode)

						n, err := w.Write(cached.Body)
						if err != nil {
							slog.Error(
								"failed to write cached response",
								"error", err,
								"cache_key", key,
								"bytes_written", n,
								"body_size", len(cached.Body),
							)
						}

						slog.Info(
							"cache response sent",
							"cache_key", key,
							"status", cached.StatusCode,
							"bytes_written", n,
							"body_size", len(cached.Body),
						)
						return
					}
				} else {
					slog.Warn(
						"failed to decode cached response",
						"error", err,
						"cache_key", key,
					)
				}
			}

			// Кэш пустой или повреждённый — идём в сервис.
			w.Header().Set("X-Cache", "MISS")

			cw := NewCachedResponseWriter(w)
			next.ServeHTTP(cw, r)

			if cw.statusCode < 200 || cw.statusCode >= 300 {
				return
			}

			body := cw.body.Bytes()
			if len(body) == 0 || !json.Valid(body) {
				return
			}

			headers := make(map[string]string)
			for name, values := range cw.Header() {
				if len(values) == 0 {
					continue
				}

				// Не сохраняем заголовки, зависящие от передачи тела.
				switch http.CanonicalHeaderKey(name) {
				case "Content-Length", "Transfer-Encoding":
					continue
				}

				headers[name] = values[0]
			}

			cached := cachepkg.CachedResponse{
				StatusCode: cw.statusCode,
				Headers:    headers,
				Body:       json.RawMessage(body),
			}

			encoded, err := json.Marshal(cached)
			if err != nil {
				slog.Error(
					"failed to serialize cached response",
					"error", err,
					"cache_key", key,
				)
				return
			}

			cfg.Cache.Set(
				r.Context(),
				key,
				string(encoded),
				cfg.TTL,
			)

			slog.Info(
				"cache response captured",
				"path", r.URL.Path,
				"status", cw.statusCode,
				"body_size", len(body),
			)
		})
	}
}
