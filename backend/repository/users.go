package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"forum_backend/model"
	"time"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

type UsersRepository interface {
	GetUser(ctx context.Context, id int64) (*model.UserInfo, error)
	CreateUser(ctx context.Context, sub *model.UserInfo) (*model.UserInfo, error)
	//GetAll(ctx context.Context, pageInt int, sizeInt int) (, error)
	UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error
	DeleteUser(ctx context.Context, id int64) error
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// READ by ID
func (ur *SQLiteUserRepository) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	var user model.UserInfo

	query := `
		SELECT id, user_name, created_at, profile_picture, last_seen, name, description
		FROM user
		WHERE id = ?
	`

	err := ur.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.UserName,
		&user.CreatedAt,
		&user.ProfilePicture,
		&user.LastSeen,
		&user.Name,
		&user.Description,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

// CREATE
func (ur *SQLiteUserRepository) CreateUser(ctx context.Context, user *model.UserInfo) (*model.UserInfo, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	user.CreatedAt = now
	user.LastSeen = now

	query := `INSERT INTO user (id, user_name, profile_picture, name, description, created_at, last_seen)
				VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := ur.db.ExecContext(ctx, query,
		user.ID,
		user.UserName,
		user.ProfilePicture,
		user.Name,
		user.Description,
		user.CreatedAt,
		user.LastSeen,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// UPDATE
func (ur *SQLiteUserRepository) UpdateUser(ctx context.Context, id int64, input model.UserUpdateInfo) error {
	query := `
		UPDATE user
		SET
			user_name = COALESCE(?, user_name),
			profile_picture = COALESCE(?, profile_picture),
			name = COALESCE(?, name),
			description = COALESCE(?, description)
		WHERE id = ?
	`

	res, err := ur.db.ExecContext(ctx, query,
		input.UserName,
		input.ProfilePicture,
		input.Name,
		input.Description,
		id,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

// DELETE
func (ur *SQLiteUserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM user WHERE id = ?`

	res, err := ur.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}
