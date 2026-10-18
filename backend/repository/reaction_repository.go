package repository

import (
	"context"
	"forum_backend/model"
)

type ReactionRepository interface {
	CreateReaction(ctx context.Context, input model.ReactionDTO) error
	UpdateReaction(ctx context.Context, input model.ReactionDTO) error
	DeleteReaction(ctx context.Context, input model.ReactionDTO) error
}
