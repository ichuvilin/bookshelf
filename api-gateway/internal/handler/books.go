package handler

import (
	"bookshelf/api-gateway/internal/proxy"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// BooksHandler обрабатывает запросы к книгам.
type BooksHandler struct {
	proxy *proxy.ServiceProxy
}

// NewBooksHandler создаёт BooksHandler.
func NewBooksHandler(p *proxy.ServiceProxy) *BooksHandler {
	return &BooksHandler{proxy: p}
}

// ListBooks проксирует получение списка книг.
func (h *BooksHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToBooks(w, r, "/api/v1/books")
}

// CreateBook проксирует создание книги.
func (h *BooksHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	h.proxy.ProxyToBooks(w, r, "/api/v1/books")
}

// GetBook проксирует получение книги по ID.
func (h *BooksHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID)
	h.proxy.ProxyToBooks(w, r, path)
}

// UpdateBook проксирует обновление книги.
func (h *BooksHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID)
	h.proxy.ProxyToBooks(w, r, path)
}

// DeleteBook проксирует удаление книги.
func (h *BooksHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	path := "/api/v1/books/" + url.PathEscape(bookID)
	h.proxy.ProxyToBooks(w, r, path)
}
