package service

import (
	"bookshelf/auth-service/internal/domain"
	"bookshelf/auth-service/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
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

	token, err := s.createAuthResponse(user)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.createAuthResponse(user)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.User, error) {
	if req.Username == "" && len(req.Username) < 3 {
		return nil, ErrInvalidUsername
	}

	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil && existingUser.ID.String() != userID {
		return nil, ErrUsernameExists
	}
	user, err := s.repo.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) createAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	token, err := s.generateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(24 * time.Hour.Seconds()),
		User:        user.ToPublic(),
	}, nil
}

func (s *UserService) generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
