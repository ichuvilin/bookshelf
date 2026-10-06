package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AuthClient struct {
	httpClient *HTTPClient
	serviceKey string
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Error     string    `json:"error"`
}

type UserPublic struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HealthResponse struct {
	Status    string           `json:"status"` // "ok", "unhealthy"
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
	Timestamp string           `json:"timestamp"`
}

type Check struct {
	Status   string `json:"status"`   // "ok", "error"
	Duration string `json:"duration"` // "2ms"
	Error    string `json:"error,omitempty"`
}

func NewAuthClient(baseURL string, timeout time.Duration, serviceKey string, maxRetries int, retryDelay time.Duration) *AuthClient {
	return &AuthClient{
		httpClient: NewHTTPClient(baseURL, timeout, 3, 10*time.Second),
		serviceKey: serviceKey,
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*VerifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	type verifyRequest struct {
		Token string `json:"token"`
	}

	resp, err := c.httpClient.Post(
		ctx,
		"/internal/v1/auth/verify",
		verifyRequest{
			Token: token,
		},
		map[string]string{
			"X-Service-Key": c.serviceKey,
		},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *AuthClient) GetUsersByIDs(ctx context.Context, ids []string) ([]UserPublic, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	type batchIDsRequest struct {
		IDs []string `json:"ids"`
	}

	resp, err := c.httpClient.Post(
		ctx,
		"/internal/v1/users/batch",
		batchIDsRequest{
			IDs: ids,
		},
		map[string]string{
			"X-Service-Key": c.serviceKey,
		})
	if err != nil {
		return nil, err
	}
	type batchUsersResponse struct {
		Users []UserPublic `json:"users"`
	}
	defer resp.Body.Close()

	var result batchUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Users, nil
}

func (c *AuthClient) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.httpClient.Get(
		ctx,
		"/health",
		nil,
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected auth-service health status: %d",
			resp.StatusCode,
		)
	}

	var result HealthResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
