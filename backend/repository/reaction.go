package repository

import (
	"context"
	"database/sql"
	"errors"
	"forum_backend/custom_err"
	"forum_backend/model"
	"log/slog"

	"github.com/mattn/go-sqlite3"
)

type SQLiteReactionRepository struct {
	db *sql.DB
}

func NewSQLiteReactionRepository(db *sql.DB) *SQLiteReactionRepository {
	return &SQLiteReactionRepository{db: db}
}

func (rr *SQLiteReactionRepository) CreateReaction(ctx context.Context, input model.ReactionDTO) error {
	/// SQL insert
	query := `INSERT INTO likes (user_id, comment_id, type_of_like)
			  VALUES (?, ?, ?)
			  ON CONFLICT(user_id, comment_id) DO NOTHING`
	_, err := rr.db.ExecContext(ctx, query,
		input.UserID,
		input.CommentID,
		input.Value,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
			return custom_err.ErrCommentNotFound
		}
		slog.Error("failed to insert like into db", "err", err)
		return custom_err.ErrCreateReaction
	}

	return nil
}

func (rr *SQLiteReactionRepository) UpdateReaction(ctx context.Context, input model.ReactionDTO) error {
	// SQL patch
	query := `
		UPDATE likes
		SET type_of_like = ?
		WHERE user_id = ? AND comment_id = ?
	`

	res, err := rr.db.ExecContext(ctx, query,
		input.Value, input.UserID, input.CommentID,
	)
	if err != nil {
		return custom_err.ErrUpdateReaction
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return custom_err.ErrGetRowsAffected
	}

	if rowsAffected == 0 {
		return custom_err.ErrReactionNotFound
	}

	return nil
}

func (rr *SQLiteReactionRepository) DeleteReaction(ctx context.Context, input model.ReactionDTO) error {
	query := `DELETE FROM likes
			  WHERE user_id = ? AND comment_id = ?`

	res, err := rr.db.ExecContext(ctx, query, input.UserID, input.CommentID)
	if err != nil {
		return custom_err.ErrDeleteReaction
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return custom_err.ErrGetRowsAffected
	}

	if rowsAffected == 0 {
		return custom_err.ErrReactionNotFound
	}

	return nil
}
