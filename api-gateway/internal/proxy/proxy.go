package proxy

import (
	"bookshelf/api-gateway/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

type ServiceProxy struct {
	authServiceURL  string
	booksServiceURL string
	client          *http.Client
}

func New(authURL, booksURL string) *ServiceProxy {
	return &ServiceProxy{
		authServiceURL:  authURL,
		booksServiceURL: booksURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *ServiceProxy) ProxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	req, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		targetURL,
		r.Body,
	)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	requestID := r.Header.Get("X-Request-ID")

	req.Header = r.Header.Clone()

	clientIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	if req.Header.Get("X-Real-IP") == "" {
		req.Header.Set("X-Real-IP", getClientIP(r))
	}

	if r.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	} else {
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		var netErr net.Error

		switch {
		case errors.Is(err, context.DeadlineExceeded),
			errors.As(err, &netErr) && netErr.Timeout():

			slog.Error("proxy request timeout",
				"target", targetURL,
				"error_type", "timeout",
				"error", err,
			)

			writeError(
				w,
				http.StatusGatewayTimeout,
				domain.ErrorResponse{
					Code:      "GATEWAY_TIMEOUT",
					Message:   "Backend service timed out",
					RequestID: requestID,
				},
			)

		case isConnectionRefused(err):
			slog.Error("proxy connection refused",
				"target", targetURL,
				"error_type", "connection_refused",
				"error", err,
			)

			writeError(
				w,
				http.StatusBadGateway,
				domain.ErrorResponse{
					Code:      "BAD_GATEWAY",
					Message:   "Backend service unavailable",
					RequestID: requestID,
				},
			)

		default:
			slog.Error("proxy request failed",
				"target", targetURL,
				"error_type", "other",
				"error", err,
			)

			writeError(
				w,
				http.StatusBadGateway,
				domain.ErrorResponse{
					Code:      "BAD_GATEWAY",
					Message:   "Backend service unavailable",
					RequestID: requestID,
				},
			)
		}

		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to copy proxy response body: %v", err)
	}
}

func (p *ServiceProxy) ProxyToAuth(w http.ResponseWriter, r *http.Request, path string) {
	targetURL := p.authServiceURL + path

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	p.ProxyRequest(w, r, targetURL)
}

func (p *ServiceProxy) ProxyToBooks(w http.ResponseWriter, r *http.Request, path string) {
	targetURL := p.booksServiceURL + path

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	p.ProxyRequest(w, r, targetURL)
}

// getClientIP возвращает IP клиента.
// Приоритет: X-Forwarded-For, X-Real-IP, RemoteAddr.
func getClientIP(r *http.Request) string {
	// X-Forwarded-For может содержать цепочку IP.
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if ip := strings.TrimSpace(strings.Split(forwardedFor, ",")[0]); ip != "" {
			return ip
		}
	}

	// Если X-Forwarded-For отсутствует, проверяем X-Real-IP.
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	// Последний вариант — адрес TCP-клиента.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	// RemoteAddr может быть без порта, например в тестах.
	return strings.Trim(r.RemoteAddr, "[]")
}

func writeError(
	w http.ResponseWriter,
	status int,
	error domain.ErrorResponse,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(error); err != nil {
		slog.Error("failed to encode error response",
			"request_id", error.RequestID,
			"error", err,
		)
	}
}

func isConnectionRefused(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}

	return errors.Is(opErr.Err, syscall.ECONNREFUSED)
}
