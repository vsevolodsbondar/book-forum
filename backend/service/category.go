package service

import (
	"context"
	"forum_backend/model"
	"forum_backend/repository"
)

type CategoryService struct {
	Repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{Repo: repo}
}

func (cs *CategoryService) GetAll(ctx context.Context) (*[]model.Category, error) {
	categories, err := cs.Repo.GetAllCategories(ctx)
	return categories, err
}
