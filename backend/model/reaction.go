package model

type ReactionDTO struct {
	UserID    int64 `json:"user_id"`
	CommentID int   `json:"comment_id"`
	Value     int   `json:"like"`
}
