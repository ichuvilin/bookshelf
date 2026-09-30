package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrBookNotFound = errors.New("book not found")

type BookRepository struct {
	db *sqlx.DB
}

func (r *BookRepository) Create(ctx context.Context, book domain.CreateBookRequest) {
	r.db.ExecContext(ctx, `INSERT INTO books (id, title, author, description, isbn, published_year, created_by) 
VALUES ($1, $2, $3, $4, $5, $6, $7)`, uuid.New(), book.Title, book.Author, book.Description, book.ISBN, book.PublishedYear, book.CreateBy)
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var book *domain.Book

	query := `
SELECT
    b.id,
       b.title,
       b.author,
       b.description,
       b.isbn,
       b.published_year,
       b.created_by,
       b.created_at,
       b.updated_at,
       count(r.*) as reviews_count,
       avg(r.rating) as average_rating
FROM books b
         JOIN reviews r ON b.id = r.book_id
WHERE b.id = $1
group by b.id
`

	err := r.db.SelectContext(ctx, &book, query, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBookNotFound
	}

	return book, nil
}

func (r *BookRepository) List(ctx context.Context, filter domain.BookFilter) ([]domain.Book, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 10
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit

	sortColumns := map[string]string{
		"title":          "title",
		"author":         "author",
		"published_year": "published_year",
		"created_at":     "created_at",
	}

	sortColumn, ok := sortColumns[filter.Sort]
	if !ok {
		sortColumn = "created_at"
	}

	order := "ASC"
	if strings.ToUpper(filter.Order) == "DESC" {
		order = "DESC"
	}

	var conditions []string
	var args []any

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")

		conditions = append(
			conditions,
			"(title ILIKE $1 OR author ILIKE $1)",
		)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM books
		%s
	`, where)

	var totalCount int

	if err := r.db.GetContext(
		ctx,
		&totalCount,
		countQuery,
		args...,
	); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			title,
			description,
			isbn,
			author,
			published_year,
			created_at,
			updated_at
		FROM books
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`,
		where,
		sortColumn,
		order,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, filter.Limit, offset)

	var books []domain.Book

	if err := r.db.SelectContext(
		ctx,
		&books,
		query,
		args...,
	); err != nil {
		return nil, 0, err
	}

	return books, totalCount, nil
}

func (r *BookRepository) Update(ctx context.Context, book domain.UpdateBookRequest) error {
	const query = `
		UPDATE books
		SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			isbn = COALESCE($3, isbn),
			author = COALESCE($4, author),
			published_year = COALESCE($5, published_year)
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		book.Title,
		book.Description,
		book.ISBN,
		book.Author,
		book.PublishedYear,
		book.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBookNotFound
	}

	return nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBookNotFound
	}

	return nil
}
