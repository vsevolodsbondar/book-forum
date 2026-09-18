package service

import (
	"context"
	"fmt"

	"auth/internal/password"
	"auth/internal/repository"
)

// UserService handles user account operations.
type UserService struct {
	userRepo *repository.UserRepository
	hasher   *password.Hasher
}

// NewUserService creates a user service.
func NewUserService(userRepo *repository.UserRepository, hasher *password.Hasher) *UserService {
	return &UserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// Register creates a user from validated, normalized input.
func (s *UserService) Register(ctx context.Context, email, username, password string) (repository.User, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return repository.User{}, fmt.Errorf("hash password: %w", err)
	}

	return s.userRepo.Create(ctx, email, username, hash)
}
