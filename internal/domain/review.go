package domain

import (
	"database/sql"
	"time"
	"uuid"
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
	ID        uuid.UUID   `json:"id"`
	BookID    string      `json:"book_id"`
	UserID    string      `json:"user_id"`
	Rating    int         `json:"rating"`
	Title     *string     `json:"title"`
	Content   string      `json:"content"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	User      UserSummary `json:"user"`
}

type CreateReviewRequest struct {
	BookID  string  `json:"book_id"`
	UserID  string  `json:"user_id"`
	Rating  int     `json:"rating,omitempty"`
	Title   *string `json:"title"`
	Content string  `json:"content,omitempty"`
}

type UpdateReviewRequest struct {
	ID      uuid.UUID `json:"id"`
	BookID  *string   `json:"book_id"`
	UserID  *string   `json:"user_id"`
	Rating  *int      `json:"rating"`
	Title   *string   `json:"title"`
	Content *string   `json:"content"`
}

type ReviewListResponse struct {
	Data       []ReviewResponse
	Pagination Pagination
}

func (r *Review) ToResponse(user *User) *ReviewResponse {
	var title *string
	if r.Title.Valid {
		title = &r.Title.String
	}
	var u UserSummary
	if user != nil {
		u = *user.ToSummary()
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
		User:      u,
	}
}
