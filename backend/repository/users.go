package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
)

type UserRepo struct {
	DB *sql.DB
}

func (UserRepo *UserRepo) GetUser(ctx context.Context, id int64) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}

func (UserRepo *UserRepo) CreateUser(ctx context.Context, sub model.UserSubmission) (model.UserInfo, error) {
	return model.UserInfo{}, nil
}
