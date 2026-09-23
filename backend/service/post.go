package service

import (
	"context"
	"forum_backend/model"
	"forum_backend/repository"
)

type PostService struct {
	Repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{
		Repo: repo,
	}
}

func (ps *PostService) GetAllPosts(ctx context.Context, dto model.SearchPostsDTO) (*[]model.Post, error) {
	err := dto.Validate()
	if err != nil {
		return nil, err
	}

	posts, err := ps.Repo.GetAll(ctx, dto)
	return posts, err
}
