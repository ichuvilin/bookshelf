package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Title       string         `json:"title" db:"title"`
	Author      string         `json:"author" db:"author"`
	Description sql.NullString `json:"description" db:"description"`
	UserID      string         `json:"user_id" db:"created_by"`
	Rating      int            `json:"-" db:"rating"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateBookRequest struct {
	Title       string  `json:"title,omitempty"`
	Author      string  `json:"author,omitempty"`
	Description *string `json:"description"`
	UserID      string  `json:"user_id"`
}

type UpdateBookRequest struct {
	Title       *string `json:"title" db:"title"`
	Description *string `json:"description" db:"description"`
	Author      *string `json:"author" db:"author"`
	UserID      string  `json:"user_id"`
}

type BookListResponse struct {
	Data       []Book     `json:"data"`
	Pagination Pagination `json:"pagination"`
}
