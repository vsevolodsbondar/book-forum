package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
)

type SQLiteUserRepository struct {
	db *sql.DB
}
type UsersRepository interface {
	GetUser(ctx context.Context, id int64) (*model.UserInfo, error)
	CreateUser(ctx context.Context, sub *model.UserSubmission) (*model.UserInfo, error)
	//GetAll(ctx context.Context, pageInt int, sizeInt int) (, error)
	//Update(ctx context.Context, ...) (..., error)
	//Delete(ctx context.Context, commentID int) error
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (ur *SQLiteUserRepository) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	return &model.UserInfo{}, nil
}

func (ur *SQLiteUserRepository) CreateUser(ctx context.Context, sub *model.UserSubmission) (*model.UserInfo, error) {
	return &model.UserInfo{}, nil
}
