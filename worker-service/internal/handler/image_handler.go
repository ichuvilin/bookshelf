package handler

import (
	"bookshelf/worker-service/internal/domain"
	"bookshelf/worker-service/internal/queue"
	"bookshelf/worker-service/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"

	"github.com/disintegration/imaging"
	"github.com/jmoiron/sqlx"
)

const (
	CoverWidth  = 400
	CoverHeight = 600
	ThumbWidth  = 100
	ThumbHeight = 150
	JPEGQuality = 85
)

type ImageHandler struct {
	storage *storage.MinIOStorage
	db      *sqlx.DB
}

func (h *ImageHandler) HandleImageCompress(body []byte) error {
	var image domain.ImageCompressMessage
	if err := json.Unmarshal(body, &image); err != nil {
		return err
	}
	ctx := context.Background()
	file, err := h.storage.GetFile(ctx, image.OriginalPath)
	if err != nil {
		return queue.NewTemporaryError(
			fmt.Errorf("get original image: %w", err),
		)
	}
	img, err := imaging.Decode(bytes.NewReader(file))
	if err != nil {
		return err
	}

	cover := imaging.Fill(
		img,
		CoverWidth,
		CoverHeight,
		imaging.Center,
		imaging.Lanczos,
	)

	// Создаём thumbnail
	thumb := imaging.Fill(
		img,
		ThumbWidth,
		ThumbHeight,
		imaging.Center,
		imaging.Lanczos,
	)

	var coverBuffer bytes.Buffer
	if err := jpeg.Encode(&coverBuffer, cover, &jpeg.Options{
		Quality: JPEGQuality,
	}); err != nil {
		return fmt.Errorf("encode cover: %w", err)
	}
	coverPath := fmt.Sprintf(
		"covers/%s/cover.jpg",
		image.BookID,
	)

	var thumbBuffer bytes.Buffer
	if err := jpeg.Encode(&thumbBuffer, thumb, &jpeg.Options{
		Quality: JPEGQuality,
	}); err != nil {
		return fmt.Errorf("encode thumbnail: %w", err)
	}
	thumbPath := fmt.Sprintf(
		"covers/%s/thumb.jpg",
		image.BookID,
	)

	if err := h.storage.UploadFile(
		ctx,
		coverPath,
		coverBuffer.Bytes(),
		"image/jpeg",
	); err != nil {
		return queue.NewTemporaryError(
			fmt.Errorf("upload cover: %w", err),
		)
	}

	if err := h.storage.UploadFile(
		ctx,
		thumbPath,
		thumbBuffer.Bytes(),
		"image/jpeg",
	); err != nil {
		return queue.NewTemporaryError(
			fmt.Errorf("upload thumbnail: %w", err),
		)
	}

	return nil
}
