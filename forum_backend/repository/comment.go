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
	GetAll(ctx context.Context, pageInt int, sizeInt int, idPost int) (model.AllComments, error)
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
func (cr *SQLiteCommentRepository) GetAll(ctx context.Context, pageInt int, sizeInt int, idPost int) (*model.AllComments, error) {
	offset := (pageInt - 1) * sizeInt
	query := `SELECT id, text, created_at, updated_at, parent_comment_id, user_id 
	FROM comment
	WHERE post_id = ?
	ORDER BY id LIMIT ? OFFSET ?`
	queryCount := `SELECT COUNT(*) FROM comment WHERE post_id = ?`
	rows, err := cr.db.QueryContext(ctx, query, idPost, sizeInt, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := model.AllComments{}
	for rows.Next() {
		var id int
		var parentCommentID, userID *int64
		var text, createdStr, updatedStr string
		if err = rows.Scan(&id, &text, &createdStr, &updatedStr, &parentCommentID, &userID); err != nil {
			return nil, err
		}
		createdDate, err := parseSQLiteTime(createdStr)
		if err != nil {
			return nil, err
		}
		updatedDate, err := parseSQLiteTime(updatedStr)
		if err != nil {
			return nil, err
		}
		comments.Comments = append(comments.Comments, model.Comment{
			ID:              int64(id),
			Text:            text,
			PostID:          int64(idPost),
			UserID:          userID,
			ParentCommentID: parentCommentID,
			CreatedAt:       createdDate,
			UpdatedAt:       updatedDate,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	countComments := 0
	if err = cr.db.QueryRowContext(ctx, queryCount, idPost).Scan(&countComments); err != nil {
		return nil, err
	}
	comments.Page = pageInt
	comments.Size = sizeInt
	comments.Total = countComments
	return &comments, nil
}
