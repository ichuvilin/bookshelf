package service

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotReviewOwner = errors.New("not your review")
	ErrInvalidRating  = errors.New("invalid rating value")
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	bookRepo   *repository.BookRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository, bookRepo *repository.BookRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, bookRepo: bookRepo}
}

func (s *ReviewService) Create(ctx context.Context, userID string, bookID string, req domain.CreateReviewRequest) (*domain.Review, error) {
	_, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	review := &domain.Review{
		ID:      uuid.New(),
		BookID:  bookID,
		UserID:  userID,
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}
	err = s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	return s.reviewRepo.GetByID(ctx, id)
}

func (s *ReviewService) ListByBook(ctx context.Context, bookID string) ([]domain.Review, error) {
	return s.reviewRepo.ListByBookID(ctx, bookID)
}

func (s *ReviewService) Update(ctx context.Context, userID string, id string, req domain.UpdateReviewRequest) (*domain.Review, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if review.UserID != userID {
		return nil, ErrNotReviewOwner
	}

	if req.Title != nil {
		review.Title = *req.Title
	}

	if req.Rating != nil {
		review.Rating = *req.Rating
	}

	if req.Content != nil {
		review.Content = *req.Content
	}

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, userID string, id string) error {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	return s.reviewRepo.Delete(ctx, id)
}
