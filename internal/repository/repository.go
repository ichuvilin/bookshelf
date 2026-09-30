package repository

import "github.com/jmoiron/sqlx"

type Repository struct {
	UserRepository   UserRepository
	BookRepository   BookRepository
	ReviewRepository ReviewRepository
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		UserRepository:   UserRepository{db: db},
		BookRepository:   BookRepository{db: db},
		ReviewRepository: ReviewRepository{db: db},
	}
}
