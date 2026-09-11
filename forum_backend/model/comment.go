package model

import "time"

type Comment struct {
	ID              int64
	Text            string
	PostID          int64
	UserID          int64
	ParentCommentID *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
