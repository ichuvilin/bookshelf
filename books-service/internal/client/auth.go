package client

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type AuthClient struct {
	httpClient *HTTPClient
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

func NewAuthClient(baseURL string, timeout time.Duration) *AuthClient {
	return &AuthClient{
		httpClient: &HTTPClient{
			client:  &http.Client{Timeout: timeout},
			baseURL: baseURL,
		},
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*VerifyResponse, error) {
	type verifyRequest struct {
		Token string `json:"token"`
	}

	body, err := json.Marshal(verifyRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(ctx, "/internal/v1/auth/verify", body, make(map[string]string))
	if err != nil {
		return nil, err
	}

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *AuthClient) GetUsersByIDs(ctx context.Context, ids []string) ([]UserPublic, error) {
	type batchIDsRequest struct {
		IDs []string `json:"ids"`
	}

	body, err := json.Marshal(batchIDsRequest{
		IDs: ids,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(ctx, "/internal/v1/users/batch", body, make(map[string]string))
	if err != nil {
		return nil, err
	}

	var result []UserPublic

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
