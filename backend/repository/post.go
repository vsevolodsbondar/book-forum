package repository

import (
	"context"
	"database/sql"
	"fmt"
	"forum_backend/custom_err"
	"forum_backend/model"
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

	initCommentIDs := []int64{}
	postIDs := []int64{}

	for _, v := range allPosts {
		postIDs = append(postIDs, v.ID)
		initCommentIDs = append(initCommentIDs, *v.InitCommentID)
	}

	mapWithComments, err := repo.findCommentsForPosts(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	mapWithLikes, err := repo.countLikesForPost(ctx, initCommentIDs)
	if err != nil {
		return nil, err
	}

	for i := range allPosts {
		allPosts[i].CommentIDs = mapWithComments[allPosts[i].ID]
		allPosts[i].Likes = mapWithLikes[int(*allPosts[i].InitCommentID)]
	}

	paginated := model.PostsPaginated{
		Posts: allPosts,
	}
	return &paginated, nil
}

func (repo *SQLitePostRepository) findAllPosts(ctx context.Context, dto model.SearchPostsDTO) ([]model.PostResultDTO, error) {
	query := `
		SELECT id, title, author_id, category_id, created_at, init_comment_id
		FROM post
	`

	var args []any

	if dto.IsSearch {
		switch *dto.SearchField {
		case "author":
			query += `WHERE author_id = ? `
			args = append(args, dto.SearchValue)

		case "category_id":
			query += `WHERE category_id = ? `
			args = append(args, dto.SearchValue)
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

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.AuthorID,
			&post.CategoryID,
			&createdAt,
			&post.InitCommentID,
		)
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

func (repo *SQLitePostRepository) countLikesForPost(ctx context.Context, commentIDs []int64) (map[int]int, error) {
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
	commentsWithLikes := map[int]int{}
	for rows.Next() {
		var commentID int
		var likeCount int

		if err := rows.Scan(&commentID, &likeCount); err != nil {
			return nil, err
		}

		commentsWithLikes[commentID] = likeCount
	}

	return commentsWithLikes, nil
}

func (repo *SQLitePostRepository) CountPosts(ctx context.Context, dto model.SearchPostsDTO) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM post
	`

	var args []any

	if dto.IsSearch {
		switch *dto.SearchField {
		case "author":
			query += `WHERE author_id = ? `
			args = append(args, *dto.SearchValue)

		case "category_id":
			query += `WHERE category_id = ? `
			args = append(args, *dto.SearchValue)
		}
	}

	var total int

	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
