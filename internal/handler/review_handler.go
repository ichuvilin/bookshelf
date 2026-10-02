package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBookReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")

	page := 1
	if value := r.URL.Query().Get("page"); value != "" {
		var err error
		page, err = strconv.Atoi(value)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "400", "Invalid page")
			return
		}
	}

	limit := 10
	if value := r.URL.Query().Get("limit"); value != "" {
		var err error
		limit, err = strconv.Atoi(value)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "400", "Invalid limit")
			return
		}
	}

	reviews, err := h.services.ReviewService.ListByBookID(r.Context(), bookID, page, limit)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	reviewId := chi.URLParam(r, "reviewId")

	response, err := h.services.ReviewService.GetByID(r.Context(), reviewId)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	log.Println(userID, bookID)

	var req domain.CreateReviewRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}
	req.UserID = userID
	req.BookID = bookID

	review, err := h.services.ReviewService.Create(r.Context(), userID, bookID, req)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyReviewed) {
			writeError(w, r, http.StatusConflict, "409", err.Error())
			return
		}
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, review)
}

func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	var req domain.UpdateReviewRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	response, err := h.services.ReviewService.Update(r.Context(), userID, reviewID, req)
	if err != nil && errors.Is(err, service.ErrNotReviewOwner) {
		writeError(w, r, http.StatusForbidden, "403", err.Error())
		return
	}
	if err != nil && errors.Is(err, repository.ErrReviewNotFound) {
		writeError(w, r, http.StatusNotFound, "404", err.Error())
		return
	}
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	reviewID := chi.URLParam(r, "reviewId")

	err := h.services.ReviewService.Delete(r.Context(), userID, reviewID)
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
