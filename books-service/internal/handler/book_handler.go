package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"errors"
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
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	userID := getUserID(r.Context())

	book, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, book.ToResponse())
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	book, err := h.svc.GetByID(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, repository.ErrBookNotFound) {
			WriteError(w, r, http.StatusNotFound, "404", err.Error())
			return
		}
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, book)
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
			WriteError(w, r, http.StatusBadRequest, "400", "Invalid page")
			return
		}
	}

	limit := 10
	if value := r.URL.Query().Get("limit"); value != "" {
		var err error
		limit, err = strconv.Atoi(value)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "400", "Invalid limit")
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
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	booksRep := make([]domain.BookResponse, 0, len(books))
	for _, book := range books {
		booksRep = append(booksRep, *book.ToResponse())
	}

	WriteJSON(w, http.StatusOK, domain.BookListResponse{
		Data:       booksRep,
		Pagination: domain.NewPagination(page, limit, total),
	})
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	var req domain.UpdateBookRequest
	err := decodeJSON(r, &req)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	userID := getUserID(r.Context())

	book, err := h.svc.Update(r.Context(), userID, bookID, req)
	if err != nil && errors.Is(err, service.ErrNotBookOwner) {
		WriteError(w, r, http.StatusForbidden, "403", err.Error())
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	userID := getUserID(r.Context())

	err := h.svc.Delete(r.Context(), userID, bookID)
	if err != nil && errors.Is(err, service.ErrNotBookOwner) {
		WriteError(w, r, http.StatusForbidden, "403", err.Error())
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}

	WriteJSON(w, http.StatusNoContent, nil)
}
