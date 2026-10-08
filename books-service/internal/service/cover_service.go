package service

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const maxCoverSize int64 = 5 * 1024 * 1024

var (
	ErrMaxCoverSize        = errors.New("file size exceeds 5 MB")
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

type CoverService struct {
	coverRepo    *repository.CoverRepository
	bookRepo     *repository.BookRepository
	minioClient  *client.MinIOClient
	rabbitClient *client.RabbitMQClient
}

func (s *CoverService) UploadCover(ctx context.Context, userID, bookID string, file io.Reader, fileSize int64, filename string) (*domain.CoverUploadResponse, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	if book.UserID != userID {
		return nil, ErrNotBookOwner
	}

	if err = validateImageFile(filename, fileSize); err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(filename))

	objectName := fmt.Sprintf(
		"covers/%s/original%s",
		bookID,
		ext,
	)

	if err = s.minioClient.UploadFile(ctx, objectName, file, fileSize, client.GetContentType(filename)); err != nil {
		return nil, err
	}

	cover := &domain.Cover{
		BookID:       bookID,
		Status:       domain.CoverStatusProcessing,
		OriginalPath: objectName,
	}

	if err = s.coverRepo.Create(ctx, cover); err != nil {
		return nil, err
	}

	if err = s.coverRepo.UpdateBookCover(ctx, bookID, domain.CoverStatusProcessing, objectName, ""); err != nil {
		return nil, err
	}

	if err = s.rabbitClient.PublishImageCompress(ctx, client.ImageCompressMessage{
		BookID:       bookID,
		CoverID:      cover.ID,
		OriginalPath: objectName,
	}); err != nil {
		return nil, err
	}

	return &domain.CoverUploadResponse{
		CoverID: cover.ID,
		Status:  domain.CoverStatusProcessing,
		Message: "",
	}, nil
}

func validateImageFile(filename string, size int64) error {
	if size > maxCoverSize {
		return ErrMaxCoverSize
	}
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return nil
	default:
		return ErrUnsupportedFileType
	}
}
