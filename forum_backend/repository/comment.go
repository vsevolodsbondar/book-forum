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
	Create(ctx context.Context, actor *model.CreateCommentRequest) (*model.Comment, error)
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
	strCreated, strUpdated := "", ""
	err = tx.QueryRowContext(ctx, query, comment.Text, comment.PostID, comment.UserID, comment.ParentCommentID).Scan(&c.ID, &c.Text, &strCreated, &strUpdated, &c.PostID, &c.ParentCommentID, &c.UserID)
	if err != nil {
		return nil, err
	}
	dateCreated, err := parseSQLiteTime(strCreated)
	if err != nil {
		return nil, err
	}
	dateUpdated, err := parseSQLiteTime(strUpdated)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = dateCreated
	c.UpdatedAt = dateUpdated
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &c, nil
}
