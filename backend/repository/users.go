package repository

import (
	"context"
	"database/sql"
	"errors"
	"forum_backend/custom_err"
	"forum_backend/model"
	"time"
)

type SQLiteUserRepository struct {
	db *sql.DB
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
			return nil, custom_err.ErrUserNotFound
		}
		return nil, custom_err.ErrGetUser
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
		return nil, custom_err.ErrCreateUser
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
		return custom_err.ErrUpdateUser
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return custom_err.ErrGetRowsAffected
	}

	if rowsAffected == 0 {
		return custom_err.ErrUserNotFound
	}

	return nil
}

// DELETE
func (ur *SQLiteUserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM user WHERE id = ?`

	res, err := ur.db.ExecContext(ctx, query, id)
	if err != nil {
		return custom_err.ErrDeleteUser
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return custom_err.ErrGetRowsAffected
	}

	if rowsAffected == 0 {
		return custom_err.ErrUserNotFound
	}

	return nil
}
