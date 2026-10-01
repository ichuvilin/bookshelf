package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"github.com/bookshelf/monolith/internal/service"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

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

	filters := domain.BookFilter{
		Search: search,
		Sort:   sort,
		Order:  order,
		Page:   page,
		Limit:  limit,
	}

	list, err := h.services.BookService.List(r.Context(), filters)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")

	book, err := h.services.BookService.GetByID(r.Context(), bookID)
	if err != nil && errors.Is(err, repository.ErrBookNotFound) {
		writeError(w, r, http.StatusNotFound, "404", err.Error())
		return
	}
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req domain.CreateBookRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	bookResponse, err := h.services.BookService.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, bookResponse)
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	var req domain.UpdateBookRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	book, err := h.services.BookService.Update(r.Context(), userID, bookID, req)
	if err != nil && errors.Is(err, service.ErrNotBookOwner) {
		writeError(w, r, http.StatusForbidden, "403", err.Error())
		return
	}
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	bookID := chi.URLParam(r, "bookId")

	err := h.services.BookService.Delete(r.Context(), userID, bookID)
	if err != nil && errors.Is(err, service.ErrNotBookOwner) {
		writeError(w, r, http.StatusForbidden, "403", err.Error())
		return
	}
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
