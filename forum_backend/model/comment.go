package model

import "time"

type Comment struct {
	ID              int64     `json:"id"`
	Text            string    `json:"text"`
	PostID          int64     `json:"post_id"`
	UserID          int64     `json:"user_id"`
	ParentCommentID *int64    `json:"parent_comment_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type CreateCommentRequest struct {
	Text            string
	PostID          int64
	UserID          int64
	ParentCommentID *int64
}
type CreateCommentBody struct {
	Text            string `json:"text"`
	ParentCommentID *int64 `json:"parent_comment_id"`
}
