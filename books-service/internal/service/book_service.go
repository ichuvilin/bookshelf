package service

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotBookOwner = errors.New("not your book")
)

type BookService struct {
	repo *repository.BookRepository
}

func NewBookService(repo *repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) Create(ctx context.Context, userID string, req domain.CreateBookRequest) (*domain.Book, error) {
	book := &domain.Book{
		ID:     uuid.New(),
		Title:  req.Title,
		Author: req.Author,
		UserID: userID,
	}

	if req.Description != nil {
		book.Description = sql.NullString{
			String: *req.Description,
			Valid:  true,
		}
	}

	if req.PublishedYear != nil {
		book.PublishedYear = sql.NullInt32{
			Int32: *req.PublishedYear,
			Valid: true,
		}
	}

	if req.ISBN != nil {
		book.ISBN = sql.NullString{
			String: *req.ISBN,
			Valid:  true,
		}
	}

	if err := s.repo.Create(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) GetByID(ctx context.Context, id string) (*domain.BookResponse, error) {
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return book.ToResponse(), err
}

func (s *BookService) List(ctx context.Context, params domain.ListParams) ([]*domain.BookResponse, int, error) {
	books, i, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	booksResponse := make([]*domain.BookResponse, 0, len(books))
	for _, book := range books {
		booksResponse = append(booksResponse, book.ToResponse())
	}
	return booksResponse, i, err
}

func (s *BookService) ListByUser(ctx context.Context, userID string, params domain.ListParams) ([]*domain.BookResponse, int, error) {
	books, i, err := s.repo.ListByUserID(ctx, userID, params)
	if err != nil {
		return nil, 0, err
	}
	booksResponse := make([]*domain.BookResponse, 0, len(books))
	for _, book := range books {
		booksResponse = append(booksResponse, book.ToResponse())
	}
	return booksResponse, i, err
}

func (s *BookService) Update(ctx context.Context, userID string, id string, req domain.UpdateBookRequest) (*domain.Book, error) {
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if book.UserID != userID {
		return nil, ErrNotBookOwner
	}

	if req.Title != nil {
		book.Title = *req.Title
	}

	if req.Description != nil {
		book.Description = sql.NullString{
			String: *req.Description,
			Valid:  true,
		}
	}

	if req.Author != nil {
		book.Author = *req.Author
	}

	if err := s.repo.Update(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookService) Delete(ctx context.Context, userID string, id string) error {
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if book.UserID != userID {
		return ErrNotBookOwner
	}

	return s.repo.Delete(ctx, id)
}
