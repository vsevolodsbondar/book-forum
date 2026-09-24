package repository

import (
	"context"
	"database/sql"
)

// HealthRepository checks database availability.
type HealthRepository struct {
	db *sql.DB
}

// NewHealthRepository creates a health repository.
func NewHealthRepository(db *sql.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

// CheckDatabase verifies that the users table can be queried.
// An empty table is healthy; the query result itself is not inspected.
func (r *HealthRepository) CheckDatabase(ctx context.Context) error {
	var exists int
	return r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users LIMIT 1)").Scan(&exists)
}
