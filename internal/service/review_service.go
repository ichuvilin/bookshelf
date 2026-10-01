package service

import (
	"context"
	"errors"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
)

var (
	ErrReviewNotFound        = errors.New("review not found")
	ErrNotReviewOwner        = errors.New("you not review owner")
	ErrAlreadyReviewed       = errors.New("already review")
	ErrInvalidRating         = errors.New("invalid rating value")
	ErrReviewContentTooShort = errors.New("review content too short")
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	bookRepo   *repository.BookRepository
	userRepo   *repository.UserRepository
}

func (s *ReviewService) Create(ctx context.Context, userID, bookID string, req domain.CreateReviewRequest) (*domain.ReviewResponse, error) {
	_, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	hasReview, err := s.reviewRepo.UserHasReviewedBook(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if hasReview {
		return nil, ErrAlreadyReviewed
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	if len(req.Content) < 10 {
		return nil, ErrReviewContentTooShort
	}

	review, err := s.reviewRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return review.ToResponse(user), nil
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, review.UserID)
	if err != nil {
		return nil, err
	}

	return review.ToResponse(user), nil
}

func (s *ReviewService) ListByBookID(ctx context.Context, bookID string, page, limit int) (*domain.ReviewListResponse, error) {
	_, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	reviews, totalCount, err := s.reviewRepo.ListByBookID(ctx, bookID, page, limit)
	if err != nil {
		return nil, err
	}

	data := make([]domain.ReviewResponse, 0, len(reviews))

	for _, v := range reviews {
		user, _ := s.userRepo.GetByID(ctx, v.UserID)
		data = append(data, *v.ToResponse(user))
	}

	return &domain.ReviewListResponse{
		Data: data,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: totalCount,
		},
	}, nil
}

func (s *ReviewService) Update(ctx context.Context, userID, reviewID string, req domain.UpdateReviewRequest) (*domain.ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.UserID != userID {
		return nil, ErrNotReviewOwner
	}

	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return nil, ErrInvalidRating
	}

	if req.Content != nil && len(*req.Content) < 10 {
		return nil, ErrReviewContentTooShort
	}

	update, err := s.reviewRepo.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return update.ToResponse(user), nil
}

func (s *ReviewService) Delete(ctx context.Context, userID, reviewID string) error {
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	return s.reviewRepo.Delete(ctx, reviewID)
}
