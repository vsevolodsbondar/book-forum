package model

type Like struct {
	ID        int64 `json:"id"`
	UserID    int64 `json:"user_id"`
	CommentID int64 `json:"comment_id"`
	Type      bool  `json:"like_type"`
}
