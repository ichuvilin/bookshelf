package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrReviewNotFound = errors.New("review not found")

type ReviewRepository struct {
	db *sqlx.DB
}

func (r *ReviewRepository) Create(ctx context.Context, req domain.CreateReviewRequest) (*domain.Review, error) {
	var review domain.Review

	err := r.db.GetContext(
		ctx,
		&review,
		`
            INSERT INTO reviews (
                id,
                book_id,
                user_id,
                rating,
                title,
                content
            )
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING *
        `,
		uuid.New(),
		req.BookID,
		req.UserID,
		req.Rating,
		req.Title,
		req.Content,
	)

	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	var review domain.Review

	err := r.db.GetContext(ctx, &review, "SELECT * FROM reviews WHERE id = $1", id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReviewNotFound
	}

	return &review, nil
}

func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID string, page, limit int) ([]domain.Review, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM reviews
		WHERE book_id = $1
	`

	if err := r.db.GetContext(
		ctx,
		&totalCount,
		countQuery,
		bookID,
	); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			id,
			book_id,
			user_id,
			rating,
			title,
			content,
			created_at,
			updated_at
		FROM reviews
		WHERE book_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var reviews []domain.Review

	if err := r.db.SelectContext(
		ctx,
		&reviews,
		query,
		bookID,
		limit,
		offset,
	); err != nil {
		return nil, 0, err
	}

	return reviews, totalCount, nil
}

func (r *ReviewRepository) Update(ctx context.Context, req domain.UpdateReviewRequest) (*domain.Review, error) {
	var review *domain.Review

	const query = `
		UPDATE reviews
		SET
			title = COALESCE($1, title),
			book_id = COALESCE($2, book_id),
			user_id = COALESCE($3, user_id),
			rating = COALESCE($4, rating),
			title = COALESCE($5, title),
			content = COALESCE($6, content)
		WHERE id = $7
		RETURNING *
	`

	err := r.db.SelectContext(ctx, &review, query, req.Title, req.BookID, req.UserID, req.Rating, req.Title, req.Content, req.ID)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM reviews WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrReviewNotFound
	}

	return nil
}

func (r *ReviewRepository) UserHasReviewedBook(ctx context.Context, userID, bookID string) (bool, error) {
	var exists bool

	err := r.db.GetContext(ctx, &exists, "select exists(select 1 from reviews where user_id = $1 and book_id = $2)", userID, bookID)
	if err != nil {
		return false, err
	}

	return exists, nil
}
