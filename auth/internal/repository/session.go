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

// ValidatedSession contains the identity returned for a valid session.
type ValidatedSession struct {
	ID        string
	UserID    int64
	Username  string
	ExpiresAt int64
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

// Validate records activity and returns the identity for a session that has not expired.
func (r *SessionRepository) Validate(ctx context.Context, tokenHash []byte, now, idleCutoff int64) (session ValidatedSession, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ValidatedSession{}, fmt.Errorf("begin session validation: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		UPDATE sessions
		SET last_seen_at = ?
		WHERE token_hash = ?
		AND expires_at > ?
		AND last_seen_at > ?
		RETURNING id, user_id, expires_at
	`, now, tokenHash, now, idleCutoff).Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return ValidatedSession{}, fmt.Errorf("validate session: %w", err)
	}

	if err := tx.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", session.UserID).Scan(&session.Username); err != nil {
		return ValidatedSession{}, fmt.Errorf("find session user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ValidatedSession{}, fmt.Errorf("commit session validation: %w", err)
	}

	return session, nil
}
