package service

import "github.com/bookshelf/monolith/internal/repository"

type Service struct {
	UserService   UserService
	BookService   BookService
	ReviewService ReviewService
}

func New(repos *repository.Repository, jwtSecret string) *Service {
	return &Service{
		UserService: UserService{
			repo:      &repos.UserRepository,
			jwtSecret: jwtSecret,
		},
		BookService: BookService{
			bookRepo: &repos.BookRepository,
			userRepo: &repos.UserRepository,
		},
		ReviewService: ReviewService{
			reviewRepo: &repos.ReviewRepository,
			bookRepo:   &repos.BookRepository,
			userRepo:   &repos.UserRepository,
		},
	}
}
