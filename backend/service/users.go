package service

import (
	"context"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/model"
	"forum_backend/repository"
)

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
		return nil, custom_err.ErrInvalidID
	}

	user, err := us.Repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	return created, nil
}

// UPDATE
func (us *UserService) UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error {
	if id < 1 {
		return custom_err.ErrInvalidID
	}

	// if err := validateUpdateInput(input); err != nil {
	// 	return fmt.Errorf("validation failed: %w", err)
	// }

	err := us.Repo.UpdateUser(ctx, id, input)
	if err != nil {
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
