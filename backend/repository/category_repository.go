package repository

import (
	"context"
	"forum_backend/model"
)

type CategoryRepository interface {
	GetAllCategories(ctx context.Context) (*[]model.Category, error)
}
