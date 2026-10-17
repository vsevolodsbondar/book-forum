package model

import (
	"fmt"
	"forum_backend/custom_err"
	"strings"
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
	ID            int64         `json:"id"`
	Title         string        `json:"title"`
	AuthorID      *int64        `json:"author_id"`
	CategoryID    *int64        `json:"category_id"`
	CategoryName  string        `json:"category_name"`
	InitCommentID *int64        `json:"initial_comment_id"`
	CreatedAt     time.Time     `json:"created_at"`
	CommentIDs    []int64       `json:"commentIDs"`
	Likes         int           `json:"likes"`
	Author        UserShortInfo `json:"author"`
}

type CreatePostRequestDTO struct {
	Title       *string `json:"title"`
	CategoryID  *int64  `json:"category_id"`
	InitComment *string `json:"init_comment_text"`
}

type CreatePostDTO struct {
	Title           *string `json:"title"`
	AuthorID        *int64  `json:"author_id"`
	CategoryID      *int64  `json:"category_id"`
	InitCommentText *string `json:"init_comment_text"`
}

type PostCreatedDTO struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	AuthorID        *int64    `json:"author_id"`
	CategoryName    string    `json:"category_name"`
	InitCommentID   *int64    `json:"initial_comment_id"`
	InitCommentText *string   `json:"init_comment_text"`
	CreatedAt       time.Time `json:"created_at"`
}

type DeletePostDTO struct {
	PostID int64
	UserID int64
}

type PatchPostDTO struct {
	PostID      int64
	UserID      int64
	NewTitle    *string `json:"new_title"`
	NewCategory *int64  `json:"new_category"`
}

type PatchResultDTO struct {
	Title      *string
	CategoryID *int64
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
	case "authorId":
		if dto.SearchValue == nil {
			return fmt.Errorf(
				"%w: author search value cannot be empty",
				custom_err.ErrInvalidInput,
			)
		}
	case "categoryId":
		if dto.SearchValue == nil {
			return fmt.Errorf(
				"%w: category_id search value cannot be empty",
				custom_err.ErrInvalidInput,
			)
		}
	case "title":
		if dto.SearchValue == nil || len(strings.TrimSpace(*dto.SearchValue)) < 3 {
			return fmt.Errorf(
				"%w: title search value must be at least 3 characters",
				custom_err.ErrInvalidInput,
			)
		}
	case "likedBy", "dislikedBy":
		if dto.SearchValue == nil {
			return fmt.Errorf(
				"%w: liked_by search value cannot be empty",
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

func (dto *PatchPostDTO) Validate() error {
	if dto.NewTitle != nil {
		if strings.TrimSpace(*dto.NewTitle) == "" {
			return fmt.Errorf("%w: title can't be empty", custom_err.ErrInvalidInput)
		}
	}

	if dto.NewCategory != nil {
		if *dto.NewCategory <= 0 {
			return fmt.Errorf("%w: category_id must be greater than 0", custom_err.ErrInvalidInput)
		}
	}

	return nil
}

func (dto *CreatePostRequestDTO) Validate() error {
	if dto.Title == nil || strings.TrimSpace(*dto.Title) == "" {
		return fmt.Errorf("%w: title is required", custom_err.ErrInvalidInput)
	}

	if dto.CategoryID == nil || *dto.CategoryID <= 0 {
		return fmt.Errorf("%w: category_id must be greater than 0", custom_err.ErrInvalidInput)
	}

	if dto.InitComment == nil || strings.TrimSpace(*dto.InitComment) == "" {
		return fmt.Errorf("%w: init_comment_text is required", custom_err.ErrInvalidInput)
	}

	return nil
}
