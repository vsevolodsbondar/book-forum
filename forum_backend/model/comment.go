package model

type CreateCommentRequest struct {
	Text            string
	PostID          int64
	ParentCommentID *int64
}
