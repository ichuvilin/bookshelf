package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	var req domain.CreateReviewRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	review, err := h.svc.Create(r.Context(), req.UserID, bookID, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, review)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	reviews, err := h.svc.ListByBook(r.Context(), bookID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")

	var req domain.UpdateReviewRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	response, err := h.svc.Update(r.Context(), req.UserID, reviewID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")

	type userRequest struct {
		UserID string `json:"user_id"`
	}

	var req userRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	err = h.svc.Delete(r.Context(), req.UserID, reviewID)
	if err != nil {
		if errors.Is(err, service.ErrNotReviewOwner) {
			writeError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
