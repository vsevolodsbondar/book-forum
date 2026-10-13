package repository

import (
	"context"
	"database/sql"
	"forum_backend/model"
)

type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) *SQLiteCategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

func (repo *SQLiteCategoryRepository) GetAllCategories(ctx context.Context) (*[]model.Category, error) {
	query := `
		SELECT id, name
		FROM category
	`
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []model.Category{}
	for rows.Next() {
		var id int64
		var name string

		err := rows.Scan(
			&id,
			&name,
		)
		if err != nil {
			return nil, err
		}

		category := model.Category{
			ID:   id,
			Name: name,
		}

		res = append(res, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &res, nil
}
