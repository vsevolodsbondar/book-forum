package service

import (
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

func (cs *CommentService) Create(comment *model.CreateCommentDTO) (*model.Comment, error) {
	var isValidated bool
	comment.Comment.Text, isValidated = helper.IsEmptyText(comment.Comment.Text)
	if !isValidated {
		return nil, fmt.Errorf("Your message is empty")
	}
	commentWritten, err := cs.repo.Create(comment)
	if err != nil {
		return nil, err
	}
	return commentWritten, nil
}

func (cs *CommentService) GetAll(comment model.GetAllCommentDTO) (*model.AllComments, error) {
	comments, err := cs.repo.GetAll(comment)
	return comments, err
}

func (cs *CommentService) Update(comment model.UpdateCommentDTO) (*model.UpdatedComment, error) {
	var isValidated bool
	comment.CommentToUpdate.Text, isValidated = helper.IsEmptyText(comment.CommentToUpdate.Text)
	if !isValidated {
		return nil, fmt.Errorf("Your message is empty")
	}
	updatedComment, err := cs.repo.Update(comment)
	return updatedComment, err
}

func (cs *CommentService) Delete(comment model.DeleteCommentDTO) error {
	err := cs.repo.Delete(comment)
	return err
}
