package repository

import (
	"context"
	"database/sql"
	"fmt"
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
	Create(comment *model.CreateCommentDTO) (*model.Comment, error)
	GetAll(comment model.GetAllCommentDTO) (*model.AllComments, error)
	Update(comment model.UpdateCommentDTO) (*model.UpdatedComment, error)
	Delete(ctx context.Context, commentID int) error
}

func (cr *SQLiteCommentRepository) Create(comment *model.CreateCommentDTO) (*model.Comment, error) {
	var postExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM post WHERE id = ?)`
	if err := cr.db.QueryRowContext(comment.Ctx, checkQuery, comment.Comment.PostID).Scan(&postExists); err != nil {
		return nil, err
	}
	if !postExists {
		return nil, fmt.Errorf("post with this id doesn't exist")
	}
	tx, err := cr.db.BeginTx(comment.Ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `INSERT INTO comment (text, post_id, user_id, parent_comment_id) VALUES (?,?,?, ?) RETURNING *;`
	var c model.Comment
	strCreated, strUpdated := "", ""
	err = tx.QueryRowContext(comment.Ctx, query, comment.Comment.Text, comment.Comment.PostID, comment.Comment.UserID, comment.Comment.ParentCommentID).Scan(&c.ID, &c.Text, &strCreated, &strUpdated, &c.PostID, &c.ParentCommentID, &c.UserID)
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

func (cr *SQLiteCommentRepository) GetAll(comment model.GetAllCommentDTO) (*model.AllComments, error) {
	offset := (comment.PageInt - 1) * comment.SizeInt
	query := `SELECT id, text, created_at, updated_at, parent_comment_id, user_id
	FROM comment
	WHERE post_id = ?
	ORDER BY id LIMIT ? OFFSET ?`
	queryCount := `SELECT COUNT(*) FROM comment WHERE post_id = ?`
	rows, err := cr.db.QueryContext(comment.Ctx, query, comment.IDPost, comment.SizeInt, offset)
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
			PostID:          int64(comment.IDPost),
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
	if err = cr.db.QueryRowContext(comment.Ctx, queryCount, comment.IDPost).Scan(&countComments); err != nil {
		return nil, err
	}
	comments.Page = comment.PageInt
	comments.Size = comment.SizeInt
	comments.Total = countComments
	return &comments, nil
}
func (cr *SQLiteCommentRepository) Update(comment model.UpdateCommentDTO) (*model.UpdatedComment, error) {
	tx, err := cr.db.BeginTx(comment.Ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `SELECT text, updated_at
	FROM comment
	WHERE id = ?`
	row := tx.QueryRow(query, comment.CommentID)
	var text, updated_at string
	err = row.Scan(&text, &updated_at)
	if err != nil {
		return nil, err
	}
	if text == comment.CommentToUpdate.Text {
		return nil, fmt.Errorf("comment wasn't changed")
	}
	updated := time.Now().UTC()
	updatedStr := updated.Format("2006-01-02 15:04:05")
	newQuery := `UPDATE comment SET text = ?, updated_at = ? WHERE id = ?`
	result, err := tx.ExecContext(comment.Ctx, newQuery, comment.CommentToUpdate.Text, updatedStr, comment.CommentID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("comment with this id doesn't exist")
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	updatedComment := model.UpdatedComment{
		ID:        int64(comment.CommentID),
		Text:      comment.CommentToUpdate.Text,
		UpdatedAt: updated,
	}
	return &updatedComment, nil
}
func (cr *SQLiteCommentRepository) Delete(ctx context.Context, commentID int) error {
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queryExists := `SELECT EXISTS(SELECT 1 FROM post WHERE init_comment_id = ?);`
	queryExistsParent := `SELECT EXISTS(SELECT 1 FROM comment WHERE parent_comment_id = ?);`
	queryReset := `UPDATE comment SET text = ?, updated_at = ?, user_id = ? WHERE id = ?`
	queryDelete := `DELETE FROM comment WHERE id = ?`
	row := tx.QueryRow(queryExists, commentID)
	isExists := 0
	err = row.Scan(&isExists)
	if err != nil {
		return err
	}
	if isExists == 1 {
		return fmt.Errorf("you not allowed to delete this comment (delete post)")
	}
	var isParent int
	row = tx.QueryRow(queryExistsParent, commentID)
	err = row.Scan(&isParent)
	if err != nil {
		return err
	}
	if isParent == 0 {
		result, err := tx.ExecContext(ctx, queryDelete, commentID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("comment with this id doesn't exist")
		}
	} else {
		text := "Deleted message"
		updated := time.Now().UTC().Format("2006-01-02 15:04:05")
		var userID *int64
		result, err := tx.ExecContext(ctx, queryReset, text, updated, userID, commentID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("comment with this id doesn't exist")
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
