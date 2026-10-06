package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	BookID    string         `json:"book_id" db:"book_id"`
	UserID    string         `json:"user_id" db:"user_id"`
	Rating    int            `json:"rating" db:"rating"`
	Title     sql.NullString `json:"title" db:"title"`
	Content   string         `json:"content" db:"content"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type ReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	BookID    string    `json:"book_id"`
	UserID    string    `json:"user_id"`
	Rating    int       `json:"rating"`
	Title     *string   `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateReviewRequest struct {
	Rating  int     `json:"rating,omitempty"`
	Title   *string `json:"title"`
	Content string  `json:"content,omitempty"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

type ReviewListResponse struct {
	Data  []*ReviewResponse `json:"data"`
	Total int               `json:"total"`
}

func (r *Review) ToResponse() *ReviewResponse {
	var title *string
	if r.Title.Valid {
		title = &r.Title.String
	}

	return &ReviewResponse{
		ID:        r.ID,
		BookID:    r.BookID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
