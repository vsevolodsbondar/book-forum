package model

import (
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
	Comments []FullComments `json:"comments"`
	Title    string         `json:"title"`
	Category string         `json:"category"`
	Page     int            `json:"page"`
	Size     int            `json:"size"`
	Total    int            `json:"total"`
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
	Comment CreateCommentRequest
	UserID  int
}
type GetAllCommentDTO struct {
	PageInt int
	SizeInt int
	IDPost  int
	UserID  int
}
type UpdateCommentDTO struct {
	CommentToUpdate CommentPatchRequest
	CommentID       int
	UserID          int
}
type DeleteCommentDTO struct {
	CommentID int
	UserID    int
}
type UserForComment struct {
	UserName string `json:"username"`
	Image    string `json:"image"`
}
type FullComments struct {
	ID              int64          `json:"id"`
	Text            string         `json:"text"`
	PostID          int64          `json:"post_id"`
	UserID          *int64         `json:"user_id"`
	ParentCommentID *int64         `json:"parent_comment_id"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Likes           int            `json:"likes"`
	Dislikes        int            `json:"dislikes"`
	User            UserForComment `json:"user_for_comment"`
}
