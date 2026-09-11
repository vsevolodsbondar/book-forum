package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
	"time"
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

func (cr *SQLiteCommentRepository) Create(ctx context.Context, comment *model.CreateCommentRequest) (*model.Comment, error) {
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `INSERT INTO comment (text, post_id, user_id, parent_comment_id) VALUES (?,?,?, ?) RETURNING *;`
	var c model.Comment
	createdAt, updatedAt := "", ""
	err = tx.QueryRowContext(ctx, query, comment.Text, comment.PostID, comment.UserID, comment.ParentCommentID).Scan(&c.ID, &c.Text, &c.PostID, &c.UserID, &c.ParentCommentID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	createdAtDate := time.Parse()
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return c, nil
}
