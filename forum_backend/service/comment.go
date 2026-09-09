package service

import (
	"context"
	"fmt"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/repository"
)

type CommentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
	}
}
func (cs *CommentService) Create(ctx context.Context, comment *model.CreateCommentRequest) (int64, error) {
	var isValidated bool
	comment.Text, isValidated = helper.IsEmptyText(comment.Text)
	if !isValidated {
		return 0, fmt.Errorf("Your message is empty")
	}
	id, err := cs.repo.Create(ctx, comment)
	if err != nil {
		return 0, err
	}
	return id, nil
}
