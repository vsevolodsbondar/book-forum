package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// ErrIdentityConflict is returned when an email or username is already in use.
var ErrIdentityConflict = errors.New("email or username already in use")

// UserRepository stores and retrieves users.
type UserRepository struct {
	db *sql.DB
}

// User contains persisted user identity fields.
type User struct {
	ID        int64
	Email     string
	Username  string
	CreatedAt int64
}

// NewUserRepository creates a user repository backed by db.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user with an already-hashed password.
func (r *UserRepository) Create(ctx context.Context, email, username, passwordHash string) (user User, err error) {
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES (?, ?, ?)
		RETURNING id, email, username, created_at
	`, email, username, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
	)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return User{}, ErrIdentityConflict
		}

		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
