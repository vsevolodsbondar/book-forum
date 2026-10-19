package repository

import (
	"context"
	m "forum_backend/model"
)

type PostRepository interface {
	GetAll(context.Context, m.SearchPostsDTO) (*m.PostsPaginated, error)
	CreatePost(context.Context, m.CreatePostDTO) (*m.PostCreatedDTO, error)
	CountPosts(context.Context, m.SearchPostsDTO) (int, error)
	DeleteByID(context.Context, m.DeletePostDTO) (bool, error)
	GetByID(context.Context, int64) (*m.PostResultDTO, error)
	ReplaceFieldsInPost(context.Context, map[string]any, int64) (*m.PatchResultDTO, error)
}
