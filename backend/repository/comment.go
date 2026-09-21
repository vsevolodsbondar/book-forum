package repository

import (
	"context"
	"database/sql"
<<<<<<< HEAD
	"fmt"
=======
	"errors"
	"fmt"
	"forum_backend/custom_err"
>>>>>>> back-auth-connection
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
<<<<<<< HEAD
	Create(ctx context.Context, actor *model.CreateCommentRequest) (*model.Comment, error)
	GetAll(ctx context.Context, pageInt int, sizeInt int, idPost int) (*model.AllComments, error)
	Update(ctx context.Context, comment model.CommentPatchRequest, commentID int) (*model.UpdatedComment, error)
	Delete(ctx context.Context, commentID int) error
}

func (cr *SQLiteCommentRepository) Create(ctx context.Context, comment *model.CreateCommentRequest) (*model.Comment, error) {
	var postExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM post WHERE id = ?)`
	if err := cr.db.QueryRowContext(ctx, checkQuery, comment.PostID).Scan(&postExists); err != nil {
		return nil, err
	}
	if !postExists {
		return nil, fmt.Errorf("post with this id doesn't exist")
=======
	Create(ctx context.Context, comment *model.CreateCommentDTO) (*model.Comment, error)
	GetAllByPostID(ctx context.Context, comment model.GetAllCommentDTO) (*model.AllComments, error)
	Update(ctx context.Context, comment model.UpdateCommentDTO) (*model.UpdatedComment, error)
	Delete(ctx context.Context, comment model.DeleteCommentDTO) error
}

func (cr *SQLiteCommentRepository) Create(ctx context.Context, comment *model.CreateCommentDTO) (*model.Comment, error) {
	var postExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM post WHERE id = ?)`
	if err := cr.db.QueryRowContext(ctx, checkQuery, comment.Comment.PostID).Scan(&postExists); err != nil {
		return nil, err
	}
	if !postExists {
		return nil, fmt.Errorf("Post with this id doesn't exist: %w", custom_err.ErrPostNotFound)
>>>>>>> back-auth-connection
	}
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := `INSERT INTO comment (text, post_id, user_id, parent_comment_id) VALUES (?,?,?, ?) RETURNING *;`
	var c model.Comment
	strCreated, strUpdated := "", ""
<<<<<<< HEAD
	err = tx.QueryRowContext(ctx, query, comment.Text, comment.PostID, comment.UserID, comment.ParentCommentID).Scan(&c.ID, &c.Text, &strCreated, &strUpdated, &c.PostID, &c.ParentCommentID, &c.UserID)
=======
	err = tx.QueryRowContext(ctx, query, comment.Comment.Text, comment.Comment.PostID, comment.UserID, comment.Comment.ParentCommentID).Scan(&c.ID, &c.Text, &strCreated, &strUpdated, &c.PostID, &c.ParentCommentID, &c.UserID)
>>>>>>> back-auth-connection
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
<<<<<<< HEAD

func (cr *SQLiteCommentRepository) GetAll(ctx context.Context, pageInt int, sizeInt int, idPost int) (*model.AllComments, error) {
	offset := (pageInt - 1) * sizeInt
	query := `SELECT id, text, created_at, updated_at, parent_comment_id, user_id
	FROM comment
	WHERE post_id = ?
	ORDER BY id LIMIT ? OFFSET ?`
	queryCount := `SELECT COUNT(*) FROM comment WHERE post_id = ?`
	rows, err := cr.db.QueryContext(ctx, query, idPost, sizeInt, offset)
=======
func (cr *SQLiteCommentRepository) GetAllByPostID(ctx context.Context, comment model.GetAllCommentDTO) (*model.AllComments, error) {
	offset := (comment.PageInt - 1) * comment.SizeInt
	query := `SELECT c.id, c.text, c.updated_at, c.parent_comment_id, c.user_id, 
	COALESCE(u.user_name, 'Deleted user') AS user_name, 
	COALESCE(u.profile_picture, '') AS profile_picture,
	(SELECT COUNT(*) FROM likes l WHERE l.comment_id = c.id AND l.type_of_like = 1) AS likes,
    (SELECT COUNT(*) FROM likes l WHERE l.comment_id = c.id AND l.type_of_like = 0) AS dislikes,
	EXISTS( SELECT 1 FROM likes l WHERE l.comment_id = c.id AND l.user_id = $1 AND l.type_of_like = 1) AS is_liked,
	EXISTS(SELECT 1 FROM likes l WHERE l.comment_id = c.id AND l.user_id = $1 AND l.type_of_like = 0) AS is_disliked
	FROM comment c
	LEFT JOIN user u ON c.user_id = u.id
	WHERE c.post_id = $2
	ORDER BY c.id LIMIT $3 OFFSET $4`
	queryCount := `SELECT COUNT(*) FROM comment WHERE post_id = ?`
	queryTitleCategory := `SELECT p.title, cat.name FROM post p 
	LEFT JOIN category cat ON p.category_id = cat.id
	WHERE p.id = ?`
	//first check title and category
	row := cr.db.QueryRowContext(ctx, queryTitleCategory, comment.IDPost)
	var title, category string
	err := row.Scan(&title, &category)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("The post with this id doesn't exist: %w", custom_err.ErrPostNotFound)
	}
	if err != nil {
		return nil, err
	}
	comments := model.AllComments{}
	comments.Category = category
	comments.Title = title
	rows, err := cr.db.QueryContext(ctx, query, comment.UserID, comment.IDPost, comment.SizeInt, offset)
>>>>>>> back-auth-connection
	if err != nil {
		return nil, err
	}
	defer rows.Close()
<<<<<<< HEAD
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
=======
	for rows.Next() {
		var id, likes, dislikes int
		var parentCommentID, userID *int64
		var text, updatedStr, userName, image string
		var isLiked, isDisliked bool
		if err = rows.Scan(&id, &text, &updatedStr, &parentCommentID, &userID, &userName, &image, &likes, &dislikes, &isLiked, &isDisliked); err != nil {
>>>>>>> back-auth-connection
			return nil, err
		}
		updatedDate, err := parseSQLiteTime(updatedStr)
		if err != nil {
			return nil, err
		}
<<<<<<< HEAD
		comments.Comments = append(comments.Comments, model.Comment{
			ID:              int64(id),
			Text:            text,
			PostID:          int64(idPost),
			UserID:          userID,
			ParentCommentID: parentCommentID,
			CreatedAt:       createdDate,
			UpdatedAt:       updatedDate,
=======
		comments.Comments = append(comments.Comments, model.FullComments{
			ID:              int64(id),
			Text:            text,
			PostID:          int64(comment.IDPost),
			UserID:          userID,
			ParentCommentID: parentCommentID,
			UpdatedAt:       updatedDate,
			Likes:           likes,
			Dislikes:        dislikes,
			User: model.UserForComment{
				UserName: userName,
				Image:    image,
			},
			IsLiked:    isLiked,
			IsDisliked: isDisliked,
>>>>>>> back-auth-connection
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	countComments := 0
<<<<<<< HEAD
	if err = cr.db.QueryRowContext(ctx, queryCount, idPost).Scan(&countComments); err != nil {
		return nil, err
	}
	comments.Page = pageInt
	comments.Size = sizeInt
	comments.Total = countComments
	return &comments, nil
}
func (cr *SQLiteCommentRepository) Update(ctx context.Context, comment model.CommentPatchRequest, commentID int) (*model.UpdatedComment, error) {
=======
	if err = cr.db.QueryRowContext(ctx, queryCount, comment.IDPost).Scan(&countComments); err != nil {
		return nil, err
	}
	comments.Page = comment.PageInt
	comments.Size = comment.SizeInt
	comments.Total = countComments
	return &comments, nil
}
func (cr *SQLiteCommentRepository) Update(ctx context.Context, comment model.UpdateCommentDTO) (*model.UpdatedComment, error) {
>>>>>>> back-auth-connection
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
<<<<<<< HEAD
	query := `SELECT text, updated_at
	FROM comment
	WHERE id = ?`
	row := tx.QueryRow(query, commentID)
	var text, updated_at string
	err = row.Scan(&text, &updated_at)
	if err != nil {
		return nil, err
	}
	if text == comment.Text {
		return nil, fmt.Errorf("comment wasn't changed")
=======
	query := `SELECT text, updated_at, user_id
	FROM comment
	WHERE id = ?`
	row := tx.QueryRowContext(ctx, query, comment.CommentID)
	var text, updated_at string
	// explicitly NULL: author removed on delete
	var userID *int
	err = row.Scan(&text, &updated_at, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("The comment with this id doesn't exist: %w", custom_err.ErrCommentNotFound)
	}
	if err != nil {
		return nil, err
	}
	if userID == nil || *userID != comment.UserID {
		return nil, fmt.Errorf("You'r not allowed to update this comment: %w", custom_err.ErrForbidden)
	}
	if text == comment.CommentToUpdate.Text {
		return nil, fmt.Errorf("Comment wasn't changed: %w", custom_err.ErrNoChange)
>>>>>>> back-auth-connection
	}
	updated := time.Now().UTC()
	updatedStr := updated.Format("2006-01-02 15:04:05")
	newQuery := `UPDATE comment SET text = ?, updated_at = ? WHERE id = ?`
<<<<<<< HEAD
	result, err := tx.ExecContext(ctx, newQuery, comment.Text, updatedStr, commentID)
=======
	result, err := tx.ExecContext(ctx, newQuery, comment.CommentToUpdate.Text, updatedStr, comment.CommentID)
>>>>>>> back-auth-connection
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
<<<<<<< HEAD
		return nil, fmt.Errorf("comment with this id doesn't exist")
=======
		return nil, fmt.Errorf("Comment with this id doesn't exist: %w", custom_err.ErrCommentNotFound)
>>>>>>> back-auth-connection
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	updatedComment := model.UpdatedComment{
<<<<<<< HEAD
		ID:        int64(commentID),
		Text:      comment.Text,
=======
		ID:        int64(comment.CommentID),
		Text:      comment.CommentToUpdate.Text,
>>>>>>> back-auth-connection
		UpdatedAt: updated,
	}
	return &updatedComment, nil
}
<<<<<<< HEAD
func (cr *SQLiteCommentRepository) Delete(ctx context.Context, commentID int) error {
=======
func (cr *SQLiteCommentRepository) Delete(ctx context.Context, comment model.DeleteCommentDTO) error {
>>>>>>> back-auth-connection
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queryExists := `SELECT EXISTS(SELECT 1 FROM post WHERE init_comment_id = ?);`
	queryExistsParent := `SELECT EXISTS(SELECT 1 FROM comment WHERE parent_comment_id = ?);`
<<<<<<< HEAD
	queryReset := `UPDATE comment SET text = ?, updated_at = ?, user_id = ? WHERE id = ?`
	queryDelete := `DELETE FROM comment WHERE id = ?`
	row := tx.QueryRow(queryExists, commentID)
=======
	queryCheckUser := `SELECT user_id FROM comment WHERE id = ?`
	queryReset := `UPDATE comment SET text = ?, updated_at = ?, user_id = ? WHERE id = ?`
	queryDelete := `DELETE FROM comment WHERE id = ?`
	row := tx.QueryRowContext(ctx, queryExists, comment.CommentID)
>>>>>>> back-auth-connection
	isExists := 0
	err = row.Scan(&isExists)
	if err != nil {
		return err
	}
	if isExists == 1 {
<<<<<<< HEAD
		return fmt.Errorf("you not allowed to delete this comment (delete post)")
	}
	var isParent int
	row = tx.QueryRow(queryExistsParent, commentID)
=======
		return fmt.Errorf("You not allowed to delete this comment (you need to delete post): %w", custom_err.ErrForbidden)
	}
	//check the rights of user
	row = tx.QueryRowContext(ctx, queryCheckUser, comment.CommentID)
	var actualUserID *int
	err = row.Scan(&actualUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("The comment with this id doesn't exist: %w", custom_err.ErrCommentNotFound)
	}
	if err != nil {
		return err
	}
	if actualUserID == nil || *actualUserID != comment.UserID {
		return fmt.Errorf("You don't have a right to delete this comment: %w", custom_err.ErrForbidden)
	}
	var isParent int
	row = tx.QueryRowContext(ctx, queryExistsParent, comment.CommentID)
>>>>>>> back-auth-connection
	err = row.Scan(&isParent)
	if err != nil {
		return err
	}
	if isParent == 0 {
<<<<<<< HEAD
		result, err := tx.ExecContext(ctx, queryDelete, commentID)
=======
		result, err := tx.ExecContext(ctx, queryDelete, comment.CommentID)
>>>>>>> back-auth-connection
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if rowsAffected == 0 {
<<<<<<< HEAD
			return fmt.Errorf("comment with this id doesn't exist")
=======
			return fmt.Errorf("Comment with this id doesn't exist: %w", custom_err.ErrCommentNotFound)
>>>>>>> back-auth-connection
		}
	} else {
		text := "Deleted message"
		updated := time.Now().UTC().Format("2006-01-02 15:04:05")
		var userID *int64
<<<<<<< HEAD
		result, err := tx.ExecContext(ctx, queryReset, text, updated, userID, commentID)
=======
		result, err := tx.ExecContext(ctx, queryReset, text, updated, userID, comment.CommentID)
>>>>>>> back-auth-connection
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if rowsAffected == 0 {
<<<<<<< HEAD
			return fmt.Errorf("comment with this id doesn't exist")
=======
			return fmt.Errorf("Comment with this id doesn't exist: %w", custom_err.ErrCommentNotFound)
>>>>>>> back-auth-connection
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
