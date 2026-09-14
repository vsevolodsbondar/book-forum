package service

import "forum_backend/repository"

type UserService struct {
	UserRepo *repository.UserRepo
}

func (UserService *UserService) GetUser(id int64)
