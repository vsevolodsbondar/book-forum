package service

import (
	"context"
	"forum_backend/model"
	"forum_backend/repository"
)

type UserService struct {
	UserRepo *repository.UserRepo
}

func (UserService *UserService) GetUser(ctx context.Context, id int64) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}

func (UserService *UserService) CreateUser(ctx context.Context, sub model.UserSubmission) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}
