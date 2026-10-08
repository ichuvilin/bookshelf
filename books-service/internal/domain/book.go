package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	Title         string         `json:"title" db:"title"`
	Author        string         `json:"author" db:"author"`
	Description   sql.NullString `json:"description" db:"description"`
	UserID        string         `json:"user_id" db:"created_by"`
	Rating        int            `json:"-" db:"rating"`
	ISBN          sql.NullString `json:"ISBN" db:"isbn"`
	PublishedYear sql.NullInt32  `json:"published_year" db:"published_year"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	CoverStatus   CoverStatus    `json:"cover_status" db:"cover_status"`
	CoverURL      string         `json:"cover_url,omitempty" db:"cover_url"`
	ThumbURL      string         `json:"thumb_url,omitempty" db:"thumb_url"`
}

type CreateBookRequest struct {
	Title         string  `json:"title,omitempty"`
	Author        string  `json:"author,omitempty"`
	Description   *string `json:"description"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int32  `json:"published_year"`
}

type UpdateBookRequest struct {
	Title       *string `json:"title" db:"title"`
	Description *string `json:"description" db:"description"`
	Author      *string `json:"author" db:"author"`
}

type BookResponse struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Description   *string   `json:"description" db:"description"`
	ISBN          *string   `json:"ISBN" db:"isbn"`
	Author        string    `json:"author" db:"author"`
	PublishedYear *int32    `json:"published_year" db:"published_year"`
	CreatedBy     string    `json:"created_by" db:"created_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func (b *Book) ToResponse() *BookResponse {
	var description *string
	var isbn *string
	var publishedYear *int32
	if b.Description.Valid {
		description = &b.Description.String
	}
	if b.ISBN.Valid {
		isbn = &b.ISBN.String
	}
	if b.PublishedYear.Valid {
		publishedYear = &b.PublishedYear.Int32
	}

	return &BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Description:   description,
		ISBN:          isbn,
		Author:        b.Author,
		PublishedYear: publishedYear,
		CreatedBy:     b.UserID,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

type BookListResponse struct {
	Data       []*BookResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
