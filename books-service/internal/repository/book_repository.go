package repository

import (
	"bookshelf/books-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrBookNotFound = errors.New("book not found")

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	return r.db.GetContext(ctx, book, "INSERT INTO books (id, title, author, description, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING *", uuid.New(), book.Title, book.Author, book.Description, book.UserID)
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	var book domain.Book

	if err := r.db.GetContext(ctx, &book, "SELECT * FROM books where id = $1", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}

	return &book, nil
}

func (r *BookRepository) List(ctx context.Context, filter domain.ListParams) ([]domain.Book, int, error) {
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
		"title":      "title",
		"author":     "author",
		"created_at": "created_at",
	}

	sortColumn, ok := sortColumns[filter.Sort]
	if !ok {
		sortColumn = "created_at"
	}

	order := "ASC"
	if strings.EqualFold(filter.Order, "DESC") {
		order = "DESC"
	}

	conditions := squirrel.And{}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"

		conditions = append(
			conditions,
			squirrel.Or{
				squirrel.Expr("title ILIKE ?", search),
				squirrel.Expr("author ILIKE ?", search),
			},
		)
	}

	countQuery, countArgs, err := squirrel.
		Select("COUNT(*)").
		From("books").
		Where(conditions).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, 0, err
	}

	var totalCount int

	if err := r.db.GetContext(
		ctx,
		&totalCount,
		countQuery,
		countArgs...,
	); err != nil {
		return nil, 0, err
	}

	query, args, err := squirrel.
		Select(
			"id",
			"title",
			"description",
			"isbn",
			"author",
			"published_year",
			"created_at",
			"updated_at",
		).
		From("books").
		Where(conditions).
		OrderBy(sortColumn + " " + order).
		Limit(uint64(filter.Limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, 0, err
	}

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

func (r *BookRepository) ListByUserID(ctx context.Context, userID string, filter domain.ListParams) ([]domain.Book, int, error) {
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
		"title":      "title",
		"author":     "author",
		"created_at": "created_at",
	}

	sortColumn, ok := sortColumns[filter.Sort]
	if !ok {
		sortColumn = "created_at"
	}

	order := "ASC"
	if strings.EqualFold(filter.Order, "DESC") {
		order = "DESC"
	}

	// Общие условия
	conditions := squirrel.And{
		squirrel.Eq{
			"creator_id": userID,
		},
	}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"

		conditions = append(
			conditions,
			squirrel.Or{
				squirrel.ILike{"title": search},
				squirrel.ILike{"author": search},
			},
		)
	}

	// COUNT
	countQuery, countArgs, err := squirrel.
		Select("COUNT(*)").
		From("books").
		Where(conditions).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, 0, err
	}

	var totalCount int

	if err := r.db.GetContext(
		ctx,
		&totalCount,
		countQuery,
		countArgs...,
	); err != nil {
		return nil, 0, err
	}

	// SELECT
	query, args, err := squirrel.
		Select(
			"id",
			"title",
			"description",
			"isbn",
			"author",
			"published_year",
			"created_at",
			"updated_at",
		).
		From("books").
		Where(conditions).
		OrderBy(sortColumn + " " + order).
		Limit(uint64(filter.Limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, 0, err
	}

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

func (r *BookRepository) Update(ctx context.Context, book *domain.Book) error {
	const query = `
        UPDATE books
        SET
            title = COALESCE($1, title),
            description = COALESCE($2, description),
            author = COALESCE($3, author),
            updated_at = NOW()
        WHERE id = $4
        RETURNING *
    `

	return r.db.GetContext(
		ctx,
		book,
		query,
		book.Title,
		book.Description,
		book.Author,
		book.ID,
	)
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
