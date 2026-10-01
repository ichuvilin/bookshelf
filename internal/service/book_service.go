package service

import (
	"context"
	"errors"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
)

var (
	ErrBookNotFound    = errors.New("book not found")
	ErrNotBookOwner    = errors.New("not your book")
	ErrBookTitleEmpty  = errors.New("book title is empty")
	ErrBookAuthorEmpty = errors.New("book author is empty")
)

type BookService struct {
	bookRepo *repository.BookRepository
	userRepo *repository.UserRepository
}

func (s *BookService) Create(ctx context.Context, userID string, req domain.CreateBookRequest) (*domain.BookResponse, error) {
	if req.Title == "" {
		return nil, ErrBookTitleEmpty
	}
	if req.Author == "" {
		return nil, ErrBookAuthorEmpty
	}

	book, err := s.bookRepo.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return book.ToResponse(), nil
}

func (s *BookService) GetByID(ctx context.Context, id string) (*domain.BookResponse, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	creator, err := s.userRepo.GetByID(ctx, book.CreatedBy)
	if err != nil {
		return nil, err
	}

	response := book.ToResponse()
	response.Creator = creator.ToSummary()

	return response, nil
}

func (s *BookService) List(ctx context.Context, filter domain.BookFilter) (*domain.BookListResponse, error) {
	list, totalCount, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	data := make([]domain.BookResponse, 0, len(list))
	for _, book := range list {
		data = append(data, *book.ToResponse())
	}

	return &domain.BookListResponse{
		Data: data,
		Pagination: domain.Pagination{
			Page:       filter.Page,
			Limit:      filter.Limit,
			Total:      totalCount,
			TotalPages: (totalCount + filter.Limit - 1) / filter.Limit,
		},
	}, nil
}

func (s *BookService) Update(ctx context.Context, userID, bookID string, req domain.UpdateBookRequest) (*domain.BookResponse, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if book.CreatedBy != userID {
		return nil, ErrNotBookOwner
	}
	book, err = s.bookRepo.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return book.ToResponse(), nil
}

func (s *BookService) Delete(ctx context.Context, userID, bookID string) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return err
	}
	if book.CreatedBy != userID {
		return ErrNotBookOwner
	}

	return s.bookRepo.Delete(ctx, bookID)
}
