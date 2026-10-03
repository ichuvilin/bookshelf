package repository

import (
	"bookshelf/books-service/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrReviewNotFound = errors.New("review not found")

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO reviews (id, book_id, user_id, rating, title, content) VALUES ($1, $2, $3, $4, $5, $6)", uuid.New(), review.BookID, review.UserID, review.Rating, review.Title, review.Content)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	var review domain.Review

	err := r.db.GetContext(ctx, &review, "SELECT * FROM reviews WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}

	return &review, nil
}

func (r *ReviewRepository) ListByBookID(ctx context.Context, bookID string) ([]domain.Review, error) {
	var reviews []domain.Review

	err := r.db.SelectContext(
		ctx,
		&reviews,
		`
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
        `,
		bookID,
	)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *domain.Review) error {
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

	result, err := r.db.ExecContext(ctx, query, review.Title, review.BookID, review.UserID, review.Rating, review.Title, review.Content, review.ID)
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
