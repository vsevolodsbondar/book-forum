package service

import (
	"context"
	"fmt"
	"forum_backend/custom_err"
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

func (cs *CommentService) Create(ctx context.Context, comment *model.CreateCommentDTO) (*model.Comment, error) {
	var isValidated bool
	comment.Comment.Text, isValidated = helper.IsEmptyText(comment.Comment.Text)
	if isValidated {
		return nil, fmt.Errorf("%w: your message is empty", custom_err.ErrInvalidInput)
	}
	commentWritten, err := cs.repo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}
	return commentWritten, nil
}

func (cs *CommentService) GetAllByPostID(ctx context.Context, comment model.GetAllCommentDTO) (*model.AllComments, error) {
	comments, err := cs.repo.GetAllByPostID(ctx, comment)
	return comments, err
}

func (cs *CommentService) Update(ctx context.Context, comment model.UpdateCommentDTO) (*model.UpdatedComment, error) {
	var isValidated bool
	comment.CommentToUpdate.Text, isValidated = helper.IsEmptyText(comment.CommentToUpdate.Text)
	if isValidated {
		return nil, fmt.Errorf("%w: your message is empty", custom_err.ErrInvalidInput)
	}
	updatedComment, err := cs.repo.Update(ctx, comment)
	return updatedComment, err
}

func (cs *CommentService) Delete(ctx context.Context, comment model.DeleteCommentDTO) error {
	err := cs.repo.Delete(ctx, comment)
	return err
}
