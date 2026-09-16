package db

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed seed.sql
var seedQuery string

func SeedDB(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin seed transaction: %w", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(seedQuery)
	if err != nil {
		return fmt.Errorf("seeding failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	fmt.Println("Database seeded")
	return nil
}
