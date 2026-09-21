package service

import (
	"context"
	"errors"
	"fmt"
	"forum_backend/client"
	"forum_backend/model"
	"forum_backend/repository"
)

// errors for input validation
var ErrInvalidID = errors.New("id must be a positive integer")
var ErrEmptyUsername = errors.New("username cannot be empty")
var ErrUsernameTooLong = errors.New("username must be at most 32 characters")
var ErrDescTooLong = errors.New("description must be at most 500 characters")

type UserService struct {
	Repo repository.UsersRepository
	Auth client.AuthInterface
}

func NewUserService(repo repository.UsersRepository, auth client.AuthInterface) *UserService {
	return &UserService{
		Repo: repo,
		Auth: auth,
	}
}

// GET
func (us *UserService) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	if id < 1 {
		return nil, ErrInvalidID
	}

	user, err := us.Repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	return user, nil
}

// CREATE
func (us *UserService) CreateUser(ctx context.Context, sub model.UserDTO) (*model.UserInfo, error) {
	authResp, err := us.Auth.RegisterUser(ctx, client.RegisterUserRequestDTO{
		Email:    sub.Email,
		Username: sub.UserName,
		Password: sub.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	info := &model.UserInfo{
		ID:             authResp.User.ID,
		UserName:       sub.UserName,
		ProfilePicture: sub.ProfilePicture,
		Name:           sub.Name,
		Description:    sub.Description,
	}

	created, err := us.Repo.CreateUser(ctx, info)
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

	// if err := validateUpdateInput(input); err != nil {
	// 	return fmt.Errorf("validation failed: %w", err)
	// }

	err := us.Repo.UpdateUser(ctx, id, input)
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

	err := us.Repo.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("user service delete: %w", err)
	}

	return nil
}

// func validateSubmission(sub model.UserDTO) error {
// 	username, empty := helper.IsEmptyText(sub.UserName)
// 	if empty {
// 		return ErrEmptyUsername
// 	}
// 	if len(username) > 32 {
// 		return ErrUsernameTooLong
// 	}
// 	if len(sub.Description) > 500 {
// 		return ErrDescTooLong
// 	}
// 	return nil
// }

// func validateUpdateInput(input model.UserUpdateInfo) error {
// 	if input.UserName != nil {
// 		username, empty := helper.IsEmptyText(*input.UserName)
// 		if empty {
// 			return ErrEmptyUsername
// 		}
// 		if len(username) > 32 {
// 			return ErrUsernameTooLong
// 		}
// 	}
// 	if input.Description != nil && len(*input.Description) > 500 {
// 		return ErrDescTooLong
// 	}
// 	return nil
// }
