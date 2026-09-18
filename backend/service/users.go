package service

import (
	"context"
	"forum_backend/model"
	"forum_backend/repository"
)

type UserService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (us *UserService) GetUser(ctx context.Context, id int64) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}

func (us *UserService) CreateUser(ctx context.Context, sub model.UserSubmission) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}
