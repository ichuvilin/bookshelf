package service

import (
	"context"
	"errors"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/bookshelf/monolith/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	if req.Username == "" || len(req.Username) < 3 {
		return nil, ErrInvalidUsername
	}
	if req.Password == "" || len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}
	if req.Email == "" {
		return nil, ErrInvalidEmail
	}

	if s.repo.EmailExists(ctx, req.Email) {
		return nil, ErrUserExists
	}

	if s.repo.UsernameExists(ctx, req.Username) {
		return nil, ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	req.Password = string(hash)

	user, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: "",
		TokenType:   "",
		ExpiresIn:   0,
		User:        user.ToPublic(),
	}, nil
}
