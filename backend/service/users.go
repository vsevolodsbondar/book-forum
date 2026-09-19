package service

import (
	"context"
	"errors"
	"fmt"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/repository"
	"math"
	"math/rand/v2"
	"strings"
)

// errors for input validation
var ErrInvalidID = errors.New("id must be a positive integer")
var ErrEmptyUsername = errors.New("username cannot be empty")
var ErrUsernameTooLong = errors.New("username must be at most 32 characters")
var ErrDescTooLong = errors.New("description must be at most 500 characters")

type UserService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// GET
func (us *UserService) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	if id < 1 {
		return nil, ErrInvalidID
	}

	user, err := us.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	return user, nil
}

// CREATE
func (us *UserService) CreateUser(ctx context.Context, sub model.UserSubmission) (*model.UserInfo, error) {
	if err := validateSubmission(sub); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	authID := rand.Int64N(math.MaxInt64) + 1

	info := &model.UserInfo{
		ID:             authID,
		UserName:       strings.TrimSpace(sub.UserName),
		ProfilePicture: strings.TrimSpace(sub.ProfilePicture),
		Name:           strings.TrimSpace(sub.Name),
		Description:    strings.TrimSpace(sub.Description),
	}

	created, err := us.repo.CreateUser(ctx, info)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

// UPDATE
func (us *UserService) UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error {
	if id < 1 {
		return ErrInvalidID
	}

	if err := validateUpdateInput(input); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err := us.repo.UpdateUser(ctx, id, input)
	if err != nil {
		return fmt.Errorf("user service update: %w", err)
	}

	return nil
}

// DELETE
func (us *UserService) DeleteUser(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrInvalidID
	}

	err := us.repo.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("user service delete: %w", err)
	}

	return nil
}

func validateSubmission(sub model.UserSubmission) error {
	username, empty := helper.IsEmptyText(sub.UserName)
	if empty {
		return ErrEmptyUsername
	}
	if len(username) > 32 {
		return ErrUsernameTooLong
	}
	if len(sub.Description) > 500 {
		return ErrDescTooLong
	}
	return nil
}

func validateUpdateInput(input model.UserUpdateInfo) error {
	if input.UserName != nil {
		username, empty := helper.IsEmptyText(*input.UserName)
		if empty {
			return ErrEmptyUsername
		}
		if len(username) > 32 {
			return ErrUsernameTooLong
		}
	}
	if input.Description != nil && len(*input.Description) > 500 {
		return ErrDescTooLong
	}
	return nil
}
