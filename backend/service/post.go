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

func (ps *PostService) GetAllPosts(ctx context.Context, dto model.SearchPostsDTO) (*model.PostsPaginated, error) {
	err := dto.Validate()
	if err != nil {
		return nil, err
	}

	total, err := ps.Repo.CountPosts(ctx, dto)
	if err != nil {
		return nil, err
	}

	totalPages := (total + dto.Limit - 1) / dto.Limit

	if dto.Page > totalPages && totalPages > 0 {
		return &model.PostsPaginated{
			Posts:      []model.PostResultDTO{},
			Page:       dto.Page,
			PageSize:   dto.Limit,
			TotalPages: totalPages,
		}, nil
	}

	postsPaginated, err := ps.Repo.GetAll(ctx, dto)
	if err != nil {
		return nil, err
	}

	postsPaginated.PageSize = dto.Limit
	postsPaginated.Page = dto.Page
	postsPaginated.TotalPages = totalPages

	return postsPaginated, err
}
