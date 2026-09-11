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
func (cs *CommentService) Create(ctx context.Context, comment *model.CreateCommentRequest) (*model.Comment, error) {
	var isValidated bool
	comment.Text, isValidated = helper.IsEmptyText(comment.Text)
	if !isValidated {
		return nil, fmt.Errorf("Your message is empty")
	}
	commentWritten, err := cs.repo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}
	return commentWritten, nil
}
func (cs *CommentService) GetAll(ctx context.Context, pageInt int, sizeInt int, idPost int) (*model.AllComments, error) {
	comments, err := cs.repo.GetAll(ctx, pageInt, sizeInt, idPost)
	return &comments, err
}
func (cs *CommentService) Update(ctx context.Context, comment model.CommentPatchRequest, commentID int) (*model.UpdatedComment, error) {
	var isValidated bool
	comment.Text, isValidated = helper.IsEmptyText(comment.Text)
	if !isValidated {
		return nil, fmt.Errorf("Your message is empty")
	}
	updatedComment, err := cs.repo.Update(ctx, comment, commentID)
	return updatedComment, err
}
func (cs *CommentService) Delete(ctx context.Context, commentID int) error {
	err := cs.repo.Delete(ctx, commentID)
	return err
}
