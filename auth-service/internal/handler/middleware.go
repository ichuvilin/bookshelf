package handler

import "net/http"

func ServiceKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Service-Key")
			if key == "" {
				writeError(
					w,
					r,
					http.StatusUnauthorized,
					"401",
					"missing service key",
				)
				return
			}

			if key != expectedKey {
				writeError(
					w,
					r,
					http.StatusForbidden,
					"403",
					"invalid service key",
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
