package handler

import (
	"errors"
	"net/http"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/service"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	response, err := h.services.UserService.Register(r.Context(), req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}
	response, err := h.services.UserService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, r, http.StatusUnauthorized, "401", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	user, err := h.services.UserService.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	writeJSON(w, http.StatusOK, user.ToPublic())
}

func (h *Handler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req domain.UpdateUserRequest
	err := decodeJSON(r, req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", "Server error")
		return
	}

	user, err := h.services.UserService.Update(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, 200, user)
}
