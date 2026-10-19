package repository

import (
	"context"
	"forum_backend/model"
)

type UserRepository interface {
	GetUser(ctx context.Context, id int64) (*model.UserInfo, error)
	CreateUser(ctx context.Context, sub *model.UserInfo) (*model.UserInfo, error)
	GetAllUsers(ctx context.Context, params model.GetAllUsersDTO) (*model.AllUsers, error)
	UpdateUser(ctx context.Context, id int64, input model.UserUpdateDTO) error
	DeleteUser(ctx context.Context, id int64) error

	CheckUsernameAvailability(ctx context.Context, username string) (bool, error)
}
