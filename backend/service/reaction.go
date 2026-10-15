package service

import (
	"context"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/repository"
)

type ReactionService struct {
	Repo repository.ReactionRepository
}

func NewReactionService(repo repository.ReactionRepository) *ReactionService {
	return &ReactionService{Repo: repo}
}

func (rs *ReactionService) CreateReaction(ctx context.Context, input model.ReactionDTO) error {
	if err := helper.ValidateValue(input.Value); err != nil {
		return err
	}
	err := rs.Repo.CreateReaction(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReactionService) UpdateReaction(ctx context.Context, input model.ReactionDTO) error {
	if err := helper.ValidateValue(input.Value); err != nil {
		return err
	}
	err := rs.Repo.UpdateReaction(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReactionService) DeleteReaction(ctx context.Context, input model.ReactionDTO) error {
	err := rs.Repo.DeleteReaction(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
