package service

import (
	"context"
	"forum_backend/custom_err"
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

	//first will check how many elements will suffice request
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

func (ps *PostService) PostMaker(ctx context.Context, dto model.CreatePostRequestDTO, id int64) (*model.PostCreatedDTO, error) {
	err := dto.Validate()
	if err != nil {
		return nil, err
	}

	createDTO := model.CreatePostDTO{
		Title:           dto.Title,
		AuthorID:        &id,
		CategoryID:      dto.CategoryID,
		InitCommentText: dto.InitComment,
	}

	res, err := ps.Repo.CreatePost(ctx, createDTO)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ps *PostService) GetByID(ctx context.Context, postID int64) (*model.PostResultDTO, error) {
	post, err := ps.Repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (ps *PostService) DeletePost(ctx context.Context, dto model.DeletePostDTO) (bool, error) {
	post, err := ps.Repo.GetByID(ctx, dto.PostID)
	if err != nil {
		return false, err
	}

	if post.AuthorID == nil || *post.AuthorID != dto.UserID {
		return false, custom_err.ErrForbidden
	}

	res, err := ps.Repo.DeleteByID(ctx, dto)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (ps *PostService) PatchPost(ctx context.Context, dto model.PatchPostDTO) (*model.PostResultDTO, error) {

	post, err := ps.Repo.GetByID(ctx, dto.PostID)
	if err != nil {
		return nil, err
	}

	return post, nil
}
