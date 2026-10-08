package handler

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CoverHandler struct {
	svc *service.CoverService
}

func NewCoverHandler(svc *service.CoverService) *CoverHandler {
	return &CoverHandler{svc: svc}
}

func (h *CoverHandler) UploadBookCover(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxCoverSize)
	if err := r.ParseMultipartForm(domain.MaxCoverSize); err != nil {
		writeError(w, r, http.StatusBadRequest, "400", err.Error())
		return
	}
	defer r.Body.Close()

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	bookID := chi.URLParam(r, "id")
	userID := getUserID(r.Context())

	cover, err := h.svc.UploadCover(r.Context(), userID, bookID, file, header.Size, header.Filename)
	if err != nil {
		if errors.As(err, &service.ErrNotBookOwner) {
			writeError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		if errors.As(err, &repository.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "404", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, cover)
}
