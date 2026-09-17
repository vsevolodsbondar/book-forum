package model

import (
	"context"
	"time"
)

type Comment struct {
	ID              int64     `json:"id"`
	Text            string    `json:"text"`
	PostID          int64     `json:"post_id"`
	UserID          *int64    `json:"user_id"`
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
type AllComments struct {
	Comments []Comment
	Page     int `json:"page"`
	Size     int `json:"size"`
	Total    int `json:"total"`
}
type CommentPatchRequest struct {
	Text string `json:"text"`
}
type UpdatedComment struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateCommentDTO struct {
	Ctx     context.Context
	Comment CreateCommentRequest
}
type GetAllCommentDTO struct {
	Ctx     context.Context
	PageInt int
	SizeInt int
	IDPost  int
}
type UpdateCommentDTO struct {
	Ctx             context.Context
	CommentToUpdate CommentPatchRequest
	CommentID       int
}
