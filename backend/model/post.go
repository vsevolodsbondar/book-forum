package model

import (
	"fmt"
	"forum_backend/custom_err"
	"time"
)

type Post struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	AuthorID      *int64    `json:"author_id"`
	CategoryID    *int64    `json:"parent_comment_id"`
	InitCommentID *int64    `json:"initial_comment_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Comments      []Comment `json:"comments"`
	Likes         []Like    `json:"likes"`
}

type SearchPostsDTO struct {
	IsSearch           bool
	SearchField        *string
	SearchValue        *string
	IsLatestPostsFirst bool
	Page               int
	Limit              int
	Offset             int
}

type PostsPaginated struct {
	Posts      []PostResultDTO `json:"posts"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
	TotalPages int             `json:"totalPages"`
}

type PostResultDTO struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	AuthorID      *int64    `json:"author_id"`
	CategoryID    *int64    `json:"parent_comment_id"`
	InitCommentID *int64    `json:"initial_comment_id"`
	CreatedAt     time.Time `json:"created_at"`
	CommentIDs    []int64   `json:"commentIDs"`
	Likes         int       `json:"likes"`
}

func (dto *SearchPostsDTO) Validate() error {
	if dto.Limit <= 0 {
		return fmt.Errorf("%w: limit must be greater than 0", custom_err.ErrInvalidInput)
	}

	if dto.Offset < 0 {
		return fmt.Errorf("%w: offset cannot be negative", custom_err.ErrInvalidInput)
	}

	if dto.SearchField == nil {
		return nil
	}

	switch *dto.SearchField {
	case "author":
		if dto.SearchValue == nil {
			return fmt.Errorf(
				"%w: author search value cannot be empty",
				custom_err.ErrInvalidInput,
			)
		}

	case "category_id":
		if dto.SearchValue == nil {
			return fmt.Errorf(
				"%w: category_id search value cannot be empty",
				custom_err.ErrInvalidInput,
			)
		}

	default:
		return fmt.Errorf(
			"%w: unsupported search field %q",
			custom_err.ErrInvalidInput,
			*dto.SearchField,
		)
	}

	return nil
}
