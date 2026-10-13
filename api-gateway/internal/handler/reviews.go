package handler

import (
	"bookshelf/api-gateway/internal/proxy"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// ReviewsHandler обрабатывает запросы к рецензиям книг.
type ReviewsHandler struct {
	proxy *proxy.ServiceProxy
}

// NewReviewsHandler создаёт ReviewsHandler.
func NewReviewsHandler(p *proxy.ServiceProxy) *ReviewsHandler {
	return &ReviewsHandler{proxy: p}
}

// ListReviews проксирует получение рецензий книги.
func (h *ReviewsHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID) + "/reviews"
	h.proxy.ProxyToBooks(w, r, path)
}

// CreateReview проксирует создание рецензии.
func (h *ReviewsHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID) + "/reviews"
	h.proxy.ProxyToBooks(w, r, path)
}

// GetReview проксирует получение рецензии по ID.
func (h *ReviewsHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	reviewID := chi.URLParam(r, "reviewId")
	if bookID == "" || reviewID == "" {
		http.Error(w, "bookId and reviewId are required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID) +
		"/reviews/" + url.PathEscape(reviewID)
	h.proxy.ProxyToBooks(w, r, path)
}

// UpdateReview проксирует обновление рецензии.
func (h *ReviewsHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	reviewID := chi.URLParam(r, "reviewId")
	if bookID == "" || reviewID == "" {
		http.Error(w, "bookId and reviewId are required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID) +
		"/reviews/" + url.PathEscape(reviewID)
	h.proxy.ProxyToBooks(w, r, path)
}

// DeleteReview проксирует удаление рецензии.
func (h *ReviewsHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	reviewID := chi.URLParam(r, "reviewId")
	if bookID == "" || reviewID == "" {
		http.Error(w, "bookId and reviewId are required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID) +
		"/reviews/" + url.PathEscape(reviewID)
	h.proxy.ProxyToBooks(w, r, path)
}
