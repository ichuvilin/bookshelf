package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bookshelf/monolith/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("email or username already exists")
)

type UserRepository struct {
	db *sqlx.DB
}

func (r *UserRepository) Create(ctx context.Context, user domain.RegisterRequest) (*domain.User, error) {
	var createdUser domain.User

	err := r.db.GetContext(ctx, &createdUser, "INSERT INTO users (id, username, email, password_hash)  VALUES ($1, $2, $3, $4) RETURNING *", uuid.New(), user.Username, user.Email, user.Password)
	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.GetContext(ctx, &user, "SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user *domain.User
	err := r.db.GetContext(ctx, &user, "SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = $1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, req domain.UpdateUserRequest) (*domain.User, error) {
	var user *domain.User

	err := r.db.GetContext(ctx, &user, "UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2", req.Username, req.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) bool {
	var exists bool

	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email)
	if err != nil {
		return false
	}

	return exists
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) bool {
	var exists bool

	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username)
	if err != nil {
		return false
	}

	return exists
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	if len(ids) == 0 {
		return map[string]*domain.User{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var users []*domain.User
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string]*domain.User, len(users))

	for _, user := range users {
		result[user.ID.String()] = user
	}

	return result, nil
}
