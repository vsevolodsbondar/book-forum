package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"forum_backend/custom_err"
	"forum_backend/model"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (ur *SQLiteUserRepository) GetAllUsers(ctx context.Context, params model.GetAllUsersDTO) (*model.AllUsers, error) {
	result := model.AllUsers{
		Users: []model.UserShortInfo{},
		Page:  params.Page,
		Size:  params.Size,
	}

	total := 0
	err := ur.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user`).Scan(&total)
	if err != nil {
		slog.Error("failed to count users", "err", err)
		return nil, custom_err.ErrGetUsers
	}

	result.TotalPages = (total + params.Size - 1) / params.Size

	offset := (params.Page - 1) * params.Size
	query := `
		SELECT id, user_name, COALESCE(profile_picture, '')
		FROM user
		ORDER BY id LIMIT ? OFFSET ?
	`

	rows, err := ur.db.QueryContext(ctx, query, params.Size, offset)
	if err != nil {
		slog.Error("failed to query users", "err", err)
		return nil, custom_err.ErrGetUsers
	}
	defer rows.Close()

	for rows.Next() {
		var user model.UserShortInfo
		if err := rows.Scan(&user.ID, &user.UserName, &user.ProfilePicture); err != nil {
			slog.Error("failed to scan user row", "err", err)
			return nil, custom_err.ErrGetUsers
		}
		result.Users = append(result.Users, user)
	}
	if err := rows.Err(); err != nil {
		slog.Error("failed to iterate users", "err", err)
		return nil, custom_err.ErrGetUsers
	}

	return &result, nil
}

// READ by ID
func (ur *SQLiteUserRepository) GetUser(ctx context.Context, id int64) (*model.UserInfo, error) {
	var user model.UserInfo

	query := `
		SELECT id, user_name, created_at, COALESCE(profile_picture, ''), last_seen, COALESCE(name, ''), COALESCE(description, '')
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

	_, err := ur.db.ExecContext(
		ctx, query,
		user.ID,
		user.UserName,
		user.ProfilePicture,
		user.Name,
		user.Description,
		user.CreatedAt,
		user.LastSeen,
	)
	if err != nil {
		slog.Error("failed to insert user into db", "err", err)
		return nil, custom_err.ErrCreateUser
	}

	return user, nil
}

// UPDATE
func (ur *SQLiteUserRepository) UpdateUser(ctx context.Context, id int64, input model.UserUpdateDTO) error {
	query := `
		UPDATE user
		SET
			user_name = COALESCE(?, user_name),
			profile_picture = COALESCE(?, profile_picture),
			name = COALESCE(?, name),
			description = COALESCE(?, description)
		WHERE id = ?
	`

	res, err := ur.db.ExecContext(
		ctx, query,
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

func (ur *SQLiteUserRepository) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	var userCount int
	err := ur.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user WHERE user_name = ?`, username).Scan(&userCount)
	return userCount == 0, err
}
