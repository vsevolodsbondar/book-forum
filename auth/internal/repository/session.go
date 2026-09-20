package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// Session contains the persisted fields of a login session.
type Session struct {
	ID         string
	UserID     int64
	TokenHash  []byte
	CreatedAt  int64
	LastSeenAt int64
	ExpiresAt  int64
}

// SessionRepository stores and retrieves login sessions.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a session repository backed by db.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create stores a new login session.
func (r *SessionRepository) Create(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (
			id,
			user_id,
			token_hash,
			created_at,
			last_seen_at,
			expires_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.CreatedAt,
		session.LastSeenAt,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}
