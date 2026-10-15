package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"forum_backend/custom_err"
	"forum_backend/model"
	m "forum_backend/model"
	"strings"
	"time"
)

type SQLitePostRepository struct {
	db *sql.DB
}

func NewSQLitePostRepository(db *sql.DB) *SQLitePostRepository {
	return &SQLitePostRepository{db: db}
}

func (repo *SQLitePostRepository) GetAll(ctx context.Context, dto model.SearchPostsDTO) (*model.PostsPaginated, error) {
	allPosts, err := repo.findAllPosts(ctx, dto)
	if err != nil {
		return nil, err
	}

	if len(allPosts) == 0 {
		return nil, custom_err.ErrPostNotFound
	}

	if err := repo.enrichPosts(ctx, allPosts); err != nil {
		return nil, err
	}

	paginated := model.PostsPaginated{
		Posts: allPosts,
	}
	return &paginated, nil
}

func (repo *SQLitePostRepository) CreatePost(ctx context.Context, dto m.CreatePostDTO) (*model.PostCreatedDTO, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	//1.check if user exists

	query := `
		INSERT INTO post (title, author_id, category_id) 
		VALUES (?,?,?)
		RETURNING id
	`

	var postID int64
	if err := tx.QueryRowContext(ctx, query, dto.Title, dto.AuthorID, dto.CategoryID).Scan(&postID); err != nil {
		return nil, err
	}

	request := model.CreateCommentRequest{
		Text:   *dto.InitCommentText,
		PostID: postID,
		UserID: *dto.AuthorID,
	}

	comment, err := repo.createCommentTx(ctx, &request, tx)
	if err != nil {
		return nil, err
	}

	post := model.PostCreatedDTO{}

	updInitCommentID := `
		UPDATE post 
		SET init_comment_id = ?
		WHERE id = ?
		RETURNING id, title, author_id, init_comment_id, created_at
	`
	var createdAt string

	err = tx.QueryRowContext(
		ctx,
		updInitCommentID,
		comment.ID,
		postID,
	).Scan(
		&post.ID,
		&post.Title,
		&post.AuthorID,
		&post.InitCommentID,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return nil, err
	}

	post.InitCommentText = &comment.Text

	// 4. Commit everything
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 5. Category name from different table
	category, err := repo.findCategoriesForPosts(ctx, []int64{*dto.CategoryID})
	if err != nil {
		return nil, err
	}
	post.CategoryName = category[*dto.CategoryID]

	return &post, nil
}

func (repo *SQLitePostRepository) DeleteByID(ctx context.Context, dto m.DeletePostDTO) (bool, error) {
	query := `
		DELETE FROM post
		WHERE id = ? AND author_id = ?
	`

	result, err := repo.db.ExecContext(
		ctx, query, dto.PostID, dto.UserID,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repo *SQLitePostRepository) GetByID(ctx context.Context, postID int64) (*m.PostResultDTO, error) {
	query := `
		SELECT id, title, author_id, category_id, created_at, init_comment_id
		FROM post
		WHERE id = ?
	`

	var post m.PostResultDTO
	var authorID sql.NullInt64
	var categoryID sql.NullInt64
	var initCommentID sql.NullInt64
	var createdAt string

	err := repo.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Title,
		&authorID,
		&categoryID,
		&createdAt,
		&initCommentID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, custom_err.ErrPostNotFound
		}
		return nil, err
	}

	if authorID.Valid {
		post.AuthorID = &authorID.Int64
	}

	if categoryID.Valid {
		post.CategoryID = &categoryID.Int64
	}

	if initCommentID.Valid {
		post.InitCommentID = &initCommentID.Int64
	}

	if err != nil {
		return nil, err
	}

	post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return nil, err
	}

	posts := []model.PostResultDTO{post}

	if err := repo.enrichPosts(ctx, posts); err != nil {
		return nil, err
	}

	return &posts[0], nil
}

func (repo *SQLitePostRepository) ReplaceFieldsInPost(ctx context.Context, fields map[string]any, postID int64) (*model.PatchResultDTO, error) {
	setFields := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	allowedFields := map[string]bool{
		"title":       true,
		"category_id": true,
	}

	for field, value := range fields {
		if !allowedFields[field] {
			return nil, custom_err.ErrInvalidInput
		}

		setFields = append(setFields, field+" = ?")
		args = append(args, value)
	}

	query := `
		UPDATE post
		SET ` + strings.Join(setFields, ", ") + `
		WHERE id = ?
		RETURNING title, category_id
	`

	args = append(args, postID)

	var res m.PatchResultDTO
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&res.Title,
		&res.CategoryID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.PatchResultDTO{}, custom_err.ErrPostNotFound
		}

		return &model.PatchResultDTO{}, err
	}

	return &res, nil
}

func (repo *SQLitePostRepository) createCommentTx(ctx context.Context, comment *model.CreateCommentRequest, tx *sql.Tx) (*model.Comment, error) {
	query := `
		INSERT INTO comment (text, post_id, user_id)
		VALUES (?,?,?) 
		RETURNING id, text;
	`

	var c model.Comment
	err := tx.QueryRowContext(
		ctx, query, comment.Text, comment.PostID, comment.UserID,
	).Scan(
		&c.ID, &c.Text,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (repo *SQLitePostRepository) findAllPosts(ctx context.Context, dto model.SearchPostsDTO) ([]model.PostResultDTO, error) {
	query := `
		SELECT id, title, author_id, category_id, created_at, init_comment_id
		FROM post
	`

	var args []any

	if dto.IsSearch {
		switch *dto.SearchField {
		case "authorId":
			query += `WHERE author_id = ? `
			args = append(args, dto.SearchValue)

		case "categoryId":
			query += `WHERE category_id = ? `
			args = append(args, dto.SearchValue)

		case "likedBy", "dislikedBy":
			reaction := 1
			if *dto.SearchField == "dislikedBy" {
				reaction = -1
			}
			query += `WHERE init_comment_id IN (
				SELECT comment_id FROM likes
				WHERE user_id = ? AND type_of_like = ?
			) `
			args = append(args, dto.SearchValue, reaction)
		}
	}

	if dto.IsLatestPostsFirst {
		query += `
			ORDER by id DESC
		`
	} else {
		query += `
			ORDER by id ASC
		`
	}

	query += `
		LIMIT ? OFFSET ?
	`

	args = append(args, dto.Limit, dto.Offset)

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []model.PostResultDTO{}

	var createdAt string
	for rows.Next() {
		var post model.PostResultDTO
		var authorID sql.NullInt64
		var categoryID sql.NullInt64
		var initCommentID sql.NullInt64

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&authorID,
			&categoryID,
			&createdAt,
			&initCommentID,
		)

		if authorID.Valid {
			post.AuthorID = &authorID.Int64
		}

		if categoryID.Valid {
			post.CategoryID = &categoryID.Int64
		}

		if initCommentID.Valid {
			post.InitCommentID = &initCommentID.Int64
		}

		if err != nil {
			return nil, err
		}

		post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repo *SQLitePostRepository) enrichPosts(ctx context.Context, posts []model.PostResultDTO) error {
	if len(posts) == 0 {
		return nil
	}

	postIDs := make([]int64, 0, len(posts))
	initCommentIDs := make([]int64, 0, len(posts))
	authorIDs := make([]int64, 0, len(posts))
	categoryIDs := make([]int64, 0, len(posts))

	for _, post := range posts {
		postIDs = append(postIDs, post.ID)

		if post.InitCommentID != nil {
			initCommentIDs = append(initCommentIDs, *post.InitCommentID)
		}
		if post.AuthorID != nil {
			authorIDs = append(authorIDs, *post.AuthorID)
		}
		if post.CategoryID != nil {
			categoryIDs = append(categoryIDs, *post.CategoryID)
		}
	}

	mapWithComments, err := repo.findCommentsForPosts(ctx, postIDs)
	if err != nil {
		return err
	}

	mapWithLikes, err := repo.countLikesForPost(ctx, initCommentIDs)
	if err != nil {
		return err
	}

	mapWithUsers, err := repo.findAuthorsForPosts(ctx, authorIDs)
	if err != nil {
		return err
	}

	mapWithCategories, err := repo.findCategoriesForPosts(ctx, categoryIDs)
	if err != nil {
		return err
	}

	for i := range posts {
		post := &posts[i]

		post.CommentIDs = mapWithComments[post.ID]

		if post.InitCommentID != nil {
			post.Likes = mapWithLikes[*post.InitCommentID]
		}

		if post.AuthorID != nil {
			post.Author = mapWithUsers[*post.AuthorID]
		}

		if post.CategoryID != nil {
			post.CategoryName = mapWithCategories[*post.CategoryID]
		}
	}

	return nil
}

func (repo *SQLitePostRepository) findCommentsForPosts(ctx context.Context, postIDs []int64) (map[int64][]int64, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]any, len(postIDs))

	for i, postID := range postIDs {
		placeholders[i] = "?"
		args[i] = postID
	}

	query := fmt.Sprintf(`
		SELECT id, post_id
		FROM comment
		WHERE post_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//key == postID, value == commentIDs for post
	postsWithComments := map[int64][]int64{}
	for rows.Next() {
		var commentID int64
		var postID int64

		if err := rows.Scan(&commentID, &postID); err != nil {
			return nil, err
		}

		postsWithComments[postID] = append(postsWithComments[postID], commentID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return postsWithComments, nil
}

func (repo *SQLitePostRepository) countLikesForPost(ctx context.Context, commentIDs []int64) (map[int64]int, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(commentIDs))
	args := make([]any, len(commentIDs))

	for i, commentID := range commentIDs {
		placeholders[i] = "?"
		args[i] = commentID
	}

	query := fmt.Sprintf(`
		SELECT
			comment_id,
			SUM(CASE WHEN type_of_like = 1 THEN 1 ELSE -1 END) AS like_count
		FROM likes
		WHERE comment_id IN (%s)
		GROUP BY comment_id
	`, strings.Join(placeholders, ","))

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//key == commentID, value == likeCount for comment
	commentsWithLikes := map[int64]int{}
	for rows.Next() {
		var commentID int64
		var likeCount int

		if err := rows.Scan(&commentID, &likeCount); err != nil {
			return nil, err
		}

		commentsWithLikes[commentID] = likeCount
	}

	return commentsWithLikes, nil
}

func (repo *SQLitePostRepository) findCategoriesForPosts(ctx context.Context, categorieIDs []int64) (map[int64]string, error) {
	if len(categorieIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(categorieIDs))
	args := make([]any, len(categorieIDs))

	for i, categoryID := range categorieIDs {
		placeholders[i] = "?"
		args[i] = categoryID
	}

	query := fmt.Sprintf(`
		SELECT id, name
		FROM category
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := map[int64]string{}
	for rows.Next() {
		var categoryID int64
		var name string

		if err := rows.Scan(&categoryID, &name); err != nil {
			return nil, err
		}

		categories[categoryID] = name
	}

	return categories, nil
}

func (repo *SQLitePostRepository) findAuthorsForPosts(ctx context.Context, authorIDs []int64) (map[int64]model.UserShortInfo, error) {
	if len(authorIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(authorIDs))
	args := make([]any, len(authorIDs))

	for i, postID := range authorIDs {
		placeholders[i] = "?"
		args[i] = postID
	}

	query := fmt.Sprintf(`
		SELECT id, user_name, profile_picture
		FROM user
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//key == userID, value == some user info for post
	users := map[int64]model.UserShortInfo{}
	for rows.Next() {
		var userId int64
		var userName string
		var profilePic sql.NullString

		if err := rows.Scan(&userId, &userName, &profilePic); err != nil {
			return nil, err
		}

		userInfo := model.UserShortInfo{
			UserName:       userName,
			ProfilePicture: profilePic.String,
		}

		users[userId] = userInfo
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (repo *SQLitePostRepository) CountPosts(ctx context.Context, dto model.SearchPostsDTO) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM post
	`
	var args []any

	if dto.IsSearch {
		switch *dto.SearchField {
		case "authorId":
			query += `WHERE author_id = ? `
			args = append(args, dto.SearchValue)

		case "categoryId":
			query += `WHERE category_id = ? `
			args = append(args, dto.SearchValue)

		case "likedBy", "dislikedBy":
			reaction := 1
			if *dto.SearchField == "dislikedBy" {
				reaction = -1
			}
			query += `WHERE init_comment_id IN (
				SELECT comment_id FROM likes
				WHERE user_id = ? AND type_of_like = ?
			) `
			args = append(args, dto.SearchValue, reaction)
		}
	}

	var total int

	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
