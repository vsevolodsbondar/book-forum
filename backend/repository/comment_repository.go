package repository

import (
	"context"
	"forum_backend/model"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.CreateCommentDTO) (*model.Comment, error)
	GetAllByPostID(ctx context.Context, comment model.GetAllCommentDTO) (*model.AllComments, error)
	Update(ctx context.Context, comment model.UpdateCommentDTO) (*model.UpdatedComment, error)
	Delete(ctx context.Context, comment model.DeleteCommentDTO) error
}
