package repository

import (
	"context"
	"forum_backend/model"
)

type PostRepository interface {
	GetAll(context.Context, model.SearchPostsDTO) (*model.PostsPaginated, error)
	CountPosts(context.Context, model.SearchPostsDTO) (int, error)
}
