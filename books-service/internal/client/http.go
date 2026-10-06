package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type HTTPClient struct {
	client     *http.Client
	baseURL    string
	maxRetries int
	retryDelay time.Duration
}

func NewHTTPClient(baseURL string, timeout time.Duration, maxRetries int, retryDelay time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    baseURL,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
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

	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+path,
			bytes.NewReader(data),
		)
		if err != nil {
			return nil, err
		}

		for name, value := range headers {
			request.Header.Set(name, value)
		}

		request.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(request)

		if err == nil {
			if !isRetryableStatus(resp.StatusCode) {
				return resp, nil
			}

			lastErr = fmt.Errorf("http status %d", resp.StatusCode)
			resp.Body.Close()
		} else {
			if !isRetryableError(err) {
				return nil, err
			}

			lastErr = err
		}

		if attempt == c.maxRetries {
			break
		}

		delay := c.retryDelay * time.Duration(1<<attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableStatus(statusCode int) bool {
	return statusCode >= 500 && statusCode <= 599
}

func isRetryableError(err error) bool {
	var netErr net.Error

	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}
