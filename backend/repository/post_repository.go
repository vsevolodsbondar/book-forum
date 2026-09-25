package repository

import (
	"context"
	m "forum_backend/model"
)

type PostRepository interface {
	GetAll(context.Context, m.SearchPostsDTO) (*m.PostsPaginated, error)
	CreatePost(context.Context, m.CreatePostDTO) (*m.PostCreatedDTO, error)
	CountPosts(context.Context, m.SearchPostsDTO) (int, error)
}
