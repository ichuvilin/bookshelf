package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type CoverRepository struct {
	db *sqlx.DB
}

func NewCoverRepository(db *sqlx.DB) *CoverRepository {
	return &CoverRepository{db: db}
}

func (r *CoverRepository) UpdateStatus(ctx context.Context, coverID string, status string, coverPath, thumbPath, errorMsg string) error {
	const query = `
		UPDATE covers
		SET
			status = $1,
			cover_path = $2,
			thumb_path = $3,
			error = $4,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		coverPath,
		thumbPath,
		errorMsg,
		coverID,
	)
	if err != nil {
		return fmt.Errorf("update cover status: %w", err)
	}

	return nil
}

func (r *CoverRepository) UpdateBookCover(ctx context.Context, bookID string, status string, coverURL, thumbURL string) error {
	const query = `
		UPDATE books
		SET
			cover_status = $1,
			cover_url = $2,
			thumbnail_url = $3,
			updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		coverURL,
		thumbURL,
		bookID,
	)
	if err != nil {
		return fmt.Errorf("update book cover: %w", err)
	}

	return nil
}
