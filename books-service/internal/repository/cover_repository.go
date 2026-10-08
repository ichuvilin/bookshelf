package repository

import (
	"bookshelf/books-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CoverRepository struct {
	db *sqlx.DB
}

func NewCoverRepository(db *sqlx.DB) *CoverRepository {
	return &CoverRepository{db: db}
}

func (r *CoverRepository) Create(ctx context.Context, cover *domain.Cover) error {
	return r.db.GetContext(
		ctx,
		cover,
		"INSERT INTO covers (id, book_id, status, original_path) VALUES ($1, $2, $3, $4) RETURNING id, book_id, status, original_path",
		uuid.New(),
		cover.BookID,
		cover.Status,
		cover.OriginalPath,
	)
}

func (r *CoverRepository) GetByBookID(ctx context.Context, bookID string) (*domain.Cover, error) {
	var cover domain.Cover

	if err := r.db.GetContext(
		ctx,
		&cover,
		"SELECT * FROM covers WHERE book_id = $1 ORDER BY created_at DESC LIMIT 1",
		bookID,
	); err != nil {
		return nil, err
	}

	return &cover, nil
}

func (r *CoverRepository) UpdateStatus(ctx context.Context, id string, status domain.CoverStatus, coverPath, thumbPath, errorMsg string) error {
	if _, err := r.db.ExecContext(
		ctx,
		`
UPDATE covers
SET status = $1, 
    cover_path = $2,
    thumb_path = $3,
    error = $4,
    completed_at = CASE
        WHEN $1 IN ('ready', 'failed') THEN now()
        ELSE completed_at
    END
WHERE id = $5
`,
		status,
		coverPath,
		thumbPath,
		errorMsg,
		id,
	); err != nil {
		return err
	}
	return nil
}

func (r *CoverRepository) UpdateBookCover(ctx context.Context, bookID string, status domain.CoverStatus, coverURL, thumbURL string) error {
	if _, err := r.db.ExecContext(
		ctx,
		`
UPDATE books
SET cover_status = $1, 
    cover_url = $2,
    thumbnail_url = $3
WHERE id = $4
`,
		status,
		coverURL,
		thumbURL,
		bookID,
	); err != nil {
		return err
	}
	return nil
}

func (r *CoverRepository) DeleteByBookID(ctx context.Context, bookID string) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM covers WHERE book_id = $1", bookID); err != nil {
		return err
	}
	return nil
}
