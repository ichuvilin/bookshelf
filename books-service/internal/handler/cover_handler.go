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

func (h *CoverHandler) GetBookCover(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	cover, err := h.svc.GetCover(r.Context(), bookID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	switch cover.Status {
	case domain.CoverStatusNone:
		writeJSON(w, http.StatusNotFound, cover)
	case domain.CoverStatusFailed:
		writeJSON(w, http.StatusInternalServerError, cover)
	case domain.CoverStatusProcessing:
		writeJSON(w, http.StatusAccepted, cover)
	case domain.CoverStatusReady:
		writeJSON(w, http.StatusOK, cover)
	}
}

func (h *CoverHandler) GetBookCoverStatus(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	coverStatus, err := h.svc.GetCoverStatus(r.Context(), bookID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, coverStatus)
}

func (h *CoverHandler) DeleteBookCover(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	userID := getUserID(r.Context())

	if err := h.svc.DeleteCover(r.Context(), userID, bookID); err != nil {
		if errors.Is(err, service.ErrNotBookOwner) {
			writeError(w, r, http.StatusForbidden, "403", err.Error())
			return
		}
		if errors.Is(err, repository.ErrBookNotFound) {
			writeError(w, r, http.StatusNotFound, "404", err.Error())
			return
		}
		writeError(w, r, http.StatusInternalServerError, "500", err.Error())
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
