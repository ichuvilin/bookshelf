package client

import (
	"context"
	"encoding/json"
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

func NewAuthClient(baseURL string, timeout time.Duration, serviceKey string) *AuthClient {
	return &AuthClient{
		httpClient: NewHTTPClient(baseURL, timeout),
		serviceKey: serviceKey,
	}
}

func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*VerifyResponse, error) {
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

	var result batchUsersResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Users, nil
}
