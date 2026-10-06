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
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	userID := getUserID(r.Context())

	review, err := h.svc.Create(r.Context(), userID, bookID, req)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, review)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "book_id")

	reviews, err := h.svc.ListByBook(r.Context(), bookID)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	reviewsResp := make([]*domain.ReviewResponse, 0, len(reviews))

	for _, review := range reviews {
		reviewsResp = append(reviewsResp, review.ToResponse())
	}

	WriteJSON(w, http.StatusOK, domain.ReviewListResponse{
		Data:  reviewsResp,
		Total: len(reviews),
	})
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")
	userID := getUserID(r.Context())

	var req domain.UpdateReviewRequest
	err := decodeJSON(r, &req)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	response, err := h.svc.Update(r.Context(), userID, reviewID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotReviewOwner) {
			WriteError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "id")
	userID := getUserID(r.Context())

	err := h.svc.Delete(r.Context(), userID, reviewID)
	if err != nil {
		if errors.Is(err, service.ErrNotReviewOwner) {
			WriteError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}
