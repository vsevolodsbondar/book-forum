package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
)

type SQLitePostRepository struct {
	db *sql.DB
}

func NewSQLitePostRepository(db *sql.DB) *SQLitePostRepository {
	return &SQLitePostRepository{db: db}
}

func (repo *SQLitePostRepository) GetAll(ctx context.Context, dto model.SearchPostsDTO) (*[]model.Post, error) {
	//filter asc or desc before return
	return &[]model.Post{}, nil
}

func (repo *SQLitePostRepository) findAllPosts(ctx context.Context, dto model.SearchPostsDTO) (map[int64]*model.Post, error) {
	query := `
		SELECT id, title, author_id, category_id, created_at, init_comment_id
		FROM post
	`

	var args []any

	if dto.IsSearch {
		switch dto.SearchField {
		case "author":
			query += `WHERE author_id = ? `
			args = append(args, dto.SearchValue)

		case "category_id":
			query += `WHERE category_id = ? `
			args = append(args, dto.SearchValue)
		}
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

	postsMap := make(map[int64]*model.Post)

	for rows.Next() {
		var post model.Post

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.AuthorID,
			&post.CategoryID,
			&post.CreatedAt,
			&post.InitCommentID,
		)
		if err != nil {
			return nil, err
		}

		postsMap[post.ID] = &post
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return postsMap, nil
}

// func (repo *SQLitePostRepository) findPostsByAuthor(ctx context.Context) (*[]model.Post, error) {
// 	query := `
// 		SELECT p.id, p.title, p.author_id, p.category_id, p.created_at, p.init_comment_id
// 		FROM post p
// 	`

// 	return &[]model.Post{}, nil
// }

// func (repo *SQLitePostRepository) findPostsByCategory(ctx context.Context) (*[]model.Post, error) {
// 	query := `
// 		SELECT p.id, p.title, p.author_id, p.category_id, p.created_at, p.init_comment_id
// 		FROM post p
// 	`

// 	return &[]model.Post{}, nil
// }
