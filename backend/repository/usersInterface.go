package repository

import (
	"context"
	"forum_backend/model"
)

type UsersRepository interface {
	GetUser(ctx context.Context, id int64) (*model.UserInfo, error)
	CreateUser(ctx context.Context, sub *model.UserInfo) (*model.UserInfo, error)
	//GetAll(ctx context.Context, pageInt int, sizeInt int) (, error)
	UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error
	DeleteUser(ctx context.Context, id int64) error
}
