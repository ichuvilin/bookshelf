package repository

import "github.com/jmoiron/sqlx"

type Repository struct {
	UserRepository UserRepository
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		UserRepository: UserRepository{db: db},
	}
}
