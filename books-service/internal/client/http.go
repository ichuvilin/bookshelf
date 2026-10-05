package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	for name, values := range headers {
		request.Header.Set(name, values)
	}
	return c.client.Do(request)
}

func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	request, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	for name, values := range headers {
		request.Header.Set(name, values)
	}
	request.Header.Set("Content-Type", "application/json")
	return c.client.Do(request)
}
