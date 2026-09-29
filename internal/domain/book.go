package domain

import (
	"database/sql"
	"time"
	"uuid"
)

type Book struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Title         string          `json:"title" db:"title"`
	Description   sql.NullString  `json:"description" db:"description"`
	ISBN          sql.NullString  `json:"ISBN" db:"isbn"`
	Author        string          `json:"author" db:"author"`
	PublishedYear sql.NullInt32   `json:"published_year" db:"published_year"`
	AverageRating sql.NullFloat64 `json:"-" db:"average_rating"`
	ReviewsCount  int             `json:"reviews_count" db:"reviews_count"`
	CreatedBy     string          `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type BookResponse struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	Title         string       `json:"title" db:"title"`
	Description   *string      `json:"description" db:"description"`
	ISBN          *string      `json:"ISBN" db:"isbn"`
	Author        string       `json:"author" db:"author"`
	PublishedYear *int32       `json:"published_year" db:"published_year"`
	AverageRating *float64     `json:"-" db:"average_rating"`
	ReviewsCount  int          `json:"reviews_count" db:"reviews_count"`
	CreatedBy     string       `json:"created_by" db:"created_by"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	Creator       *UserSummary `json:"creator,omitempty"`
}

type CreateBookRequest struct {
	Title         string  `json:"title,omitempty"`
	Author        string  `json:"author,omitempty"`
	Description   *string `json:"description"`
	ISBN          *int32  `json:"isbn"`
	PublishedYear *int32  `json:"published_year"`
}

type UpdateBookRequest struct {
	Title         *string `json:"title" db:"title"`
	Description   *string `json:"description" db:"description"`
	ISBN          *int32  `json:"ISBN" db:"isbn"`
	Author        *string `json:"author" db:"author"`
	PublishedYear *int32  `json:"published_year" db:"published_year"`
}

type BookFilter struct {
	Search string
	Sort   string
	Order  string
	Page   int
	Limit  int
}

type BookListResponse struct {
	Data       []BookResponse
	Pagination Pagination
}

func (b *Book) ToResponse() BookResponse {
	var description *string
	var isbn *string
	var publishedYear *int32
	var averageRating *float64
	if b.Description.Valid {
		description = &b.Description.String
	}
	if b.ISBN.Valid {
		isbn = &b.ISBN.String
	}
	if b.PublishedYear.Valid {
		publishedYear = &b.PublishedYear.Int32
	}
	if b.AverageRating.Valid {
		averageRating = &b.AverageRating.Float64
	}

	return BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Description:   description,
		ISBN:          isbn,
		Author:        b.Author,
		PublishedYear: publishedYear,
		AverageRating: averageRating,
		ReviewsCount:  b.ReviewsCount,
		CreatedBy:     b.CreatedBy,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
		Creator:       nil,
	}
}
