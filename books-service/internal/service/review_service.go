package service

import (
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"context"
	"errors"
)

var (
	ErrNotReviewOwner = errors.New("not your review")
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(repo *repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) Create(ctx context.Context, userID string, bookID string, req domain.CreateReviewRequest) (*domain.Review, error) {
	review := &domain.Review{
		BookID:  req.BookID,
		UserID:  userID,
		Rating:  req.Rating,
		Title:   req.Title,
		Content: req.Content,
	}
	err := s.repo.Create(ctx, review)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ReviewService) ListByBook(ctx context.Context, bookID string) ([]domain.Review, error) {
	return s.repo.ListByBookID(ctx, bookID)
}

func (s *ReviewService) Update(ctx context.Context, userID string, id string, req domain.UpdateReviewRequest) (*domain.Review, error) {
	review, err := s.repo.GetByID(ctx, id)
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

	if err := s.repo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) Delete(ctx context.Context, userID string, id string) error {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	return s.repo.Delete(ctx, id)
}
