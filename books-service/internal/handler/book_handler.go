package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	svc *service.BookService
}

func NewBookHandler(svc *service.BookService) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateBookRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	bookResponse, err := h.svc.Create(r.Context(), req.UserID, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, bookResponse)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	book, err := h.svc.GetByID(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "404", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
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

	filters := domain.ListParams{
		Search: search,
		Sort:   sort,
		Order:  order,
		Page:   page,
		Limit:  limit,
	}

	books, total, err := h.svc.List(r.Context(), filters)
	if err != nil {
		log.Println(err)
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.BookListResponse{
		Data:       books,
		Pagination: domain.NewPagination(page, limit, total),
	})
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	var req domain.UpdateBookRequest
	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	book, err := h.svc.Update(r.Context(), req.UserID, bookID, req)
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

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	type userRequest struct {
		UserID string `json:"user_id"`
	}

	var req userRequest

	err := decodeJSON(r, &req)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	err = h.svc.Delete(r.Context(), req.UserID, bookID)
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
