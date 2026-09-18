package service

import (
	"context"
	"fmt"
	"forum_backend/model"
	"forum_backend/repository"
	"math/rand/v2"
)

type UserService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (us *UserService) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	user, err := us.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context, sub model.UserSubmission) (*model.UserInfo, error) {
	//HERE WILL BE A REQUEST TO AUTH. MOCK FOR NOW
	authID := rand.Int64()

	info := &model.UserInfo{
		ID:             authID,
		UserName:       sub.UserName,
		ProfilePicture: sub.ProfilePicture,
		Name:           sub.Name,
		Description:    sub.Description,
	}

	created, err := us.repo.CreateUser(ctx, info)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

// UPD
func (us *UserService) UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error {
	if id < 1 {
		return fmt.Errorf("invalid user id")
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
		return fmt.Errorf("invalid user id")
	}

	err := us.repo.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("user service delete: %w", err)
	}

	return nil
}
