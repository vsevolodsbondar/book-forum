package repository

import (
	"context"
	"forum_backend/model"
)

type PostRepository interface {
	GetAll(context.Context, model.SearchPostsDTO) (*[]model.Post, error)
}
