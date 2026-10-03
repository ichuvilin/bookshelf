package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BookID    string    `json:"book_id" db:"book_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateReviewRequest struct {
	BookID  string `json:"book_id"`
	UserID  string `json:"user_id"`
	Rating  int    `json:"rating,omitempty"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

type UpdateReviewRequest struct {
	BookID  *string `json:"book_id"`
	UserID  *string `json:"user_id"`
	Rating  *int    `json:"rating"`
	Title   *string `json:"title"`
	Content *string `json:"content"`
}
