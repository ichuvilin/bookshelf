package service

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/domain"
	"bookshelf/books-service/internal/repository"
	"context"
	"database/sql"
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
	authClient *client.AuthClient
}

func NewReviewService(reviewRepo *repository.ReviewRepository, bookRepo *repository.BookRepository, authClient *client.AuthClient) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, bookRepo: bookRepo, authClient: authClient}
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
		Content: req.Content,
	}
	if req.Title != nil {
		review.Title = sql.NullString{
			String: *req.Title,
			Valid:  true,
		}
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

func (s *ReviewService) ListByBook(ctx context.Context, bookID string) ([]*domain.ReviewResponse, error) {
	reviews, err := s.reviewRepo.ListByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	userIDS := make([]string, 0, len(reviews))
	for _, review := range reviews {
		userIDS = append(userIDS, review.UserID)
	}

	users, err := s.authClient.GetUsersByIDs(ctx, userIDS)
	if err != nil {
		return nil, err
	}
	idToSum := make(map[string]*client.UserPublic)
	for _, user := range users {
		idToSum[user.ID] = &user
	}

	reviewsResp := make([]*domain.ReviewResponse, 0, len(reviews))

	for _, review := range reviews {
		reviewsResp = append(reviewsResp, review.ToResponse(idToSum[review.UserID]))
	}
	return reviewsResp, err
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
		review.Title = sql.NullString{
			String: *req.Title,
			Valid:  true,
		}
	}

	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, ErrInvalidRating
		}
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
