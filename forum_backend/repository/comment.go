package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
)

type SQLiteCommentRepository struct {
	db *sql.DB
}

func NewSQLiteCommentRepository(db *sql.DB) *SQLiteCommentRepository {
	return &SQLiteCommentRepository{db: db}
}

type CommentRepository interface {
	Create(ctx context.Context, actor *model.CreateCommentRequest) (int64, error)
	// GetAll(moviesFlag bool, page int, size int, pagination bool) (model.PaginatedCommentResponse, error)
	// Update(id int, actor model.CommentPatchRequest) (model.Comment, error)
	// Delete(id int, force bool) (int64, error)
}

func (c *SQLiteCommentRepository) Create(ctx context.Context, comment *model.CreateCommentRequest) (int64, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	query := `INSERT INTO comment (text, post_id, parent_comment_id) VALUES (?,?,?);`
	result, err := tx.ExecContext(ctx, query, comment.Text, comment.PostID, comment.ParentCommentID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}
