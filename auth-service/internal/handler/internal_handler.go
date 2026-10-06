package handler

import (
	"bookshelf/auth-service/internal/service"
	"net/http"
	"time"
)

type InternalHandler struct {
	svc *service.UserService
}

func NewInternalHandler(svc *service.UserService) *InternalHandler {
	return &InternalHandler{svc: svc}
}

type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Error     string    `json:"error"`
}

func (h *InternalHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	type tokenRequest struct {
		Token string `json:"token"`
	}

	var req tokenRequest

	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	token, err := h.svc.ValidateToken(req.Token)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, VerifyResponse{
			Valid:     false,
			UserID:    "",
			ExpiresAt: time.Time{},
			Error:     err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, VerifyResponse{
		Valid:     true,
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
		Error:     "",
	})
}
