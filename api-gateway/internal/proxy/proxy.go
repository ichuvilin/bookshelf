package proxy

import (
	"io"
	"log"
	"net/http"
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
	req.Header = r.Header.Clone()

	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		http.Error(w, "Backend service unavailable", http.StatusBadGateway)
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
