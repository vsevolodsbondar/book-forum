package service

import (
	"context"

	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/repository"
)

type UserService struct {
	Repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository, auth client.AuthInterface) *UserService {
	return &UserService{Repo: repo}
}

func (us *UserService) GetAllUsers(ctx context.Context, params model.GetAllUsersDTO) (*model.AllUsers, error) {
	usersInfo, err := us.Repo.GetAllUsers(ctx, params)
	if err != nil {
		return nil, err
	}
	return usersInfo, err
}

// GET
func (us *UserService) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	if id < 1 {
		return nil, custom_err.ErrInvalidID
	}

	user, err := us.Repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CREATE
func (us *UserService) CreateUser(ctx context.Context, input model.UserDTO, authResp client.RegisterUserResponseDTO) (*model.UserInfo, error) {
	info := &model.UserInfo{
		ID:             authResp.User.ID,
		UserName:       input.UserName,
		ProfilePicture: input.ProfilePicture,
		Name:           input.Name,
		Description:    input.Description,
	}

	created, err := us.Repo.CreateUser(ctx, info)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// UPDATE
func (us *UserService) UpdateUser(ctx context.Context, id int64, input model.UserUpdateDTO) error {
	if id < 1 {
		return custom_err.ErrInvalidID
	}

	if err := validateUpdateInput(us, ctx, id, input); err != nil {
		return err
	}

	if err := us.Repo.UpdateUser(ctx, id, input); err != nil {
		return err
	}

	return nil
}

// DELETE
func (us *UserService) DeleteUser(ctx context.Context, id int64) error {
	if id < 1 {
		return custom_err.ErrInvalidID
	}

	err := us.Repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func validateSubmission(input model.UserDTO) error {
	username, empty := helper.IsEmptyText(input.UserName)
	if empty {
		return custom_err.ErrEmptyUsername
	}
	if len(username) > 30 {
		return custom_err.ErrUsernameTooLong
	}
	if len(input.Description) > 500 {
		return custom_err.ErrDescTooLong
	}
	return nil
}

func validateUpdateInput(us *UserService, ctx context.Context, id int64, input model.UserUpdateDTO) error {
	user, err := us.Repo.GetUser(ctx, id)
	if err != nil {
		return err
	}

	if input.UserName != nil {
		if err := helper.ValidateUsername(*input.UserName); err != nil {
			return err
		}

		if *input.UserName == user.UserName {
			return custom_err.ErrNoChange
		}

		if free, err := us.Repo.CheckUsernameAvailability(ctx, *input.UserName); err != nil {
			return err
		} else if !free {
			return custom_err.ErrConflict
		}
	}


	if input.Description != nil && len(*input.Description) > 500 {
		return custom_err.ErrDescTooLong
	}

	return nil
}
