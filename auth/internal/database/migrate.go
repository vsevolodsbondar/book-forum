package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed migrations/*.sql
var files embed.FS

// Migrate applies pending embedded SQL migrations in filename order.
// Each migration and its history entry are committed in one transaction.
func Migrate(db *sql.DB) error {
	if err := createMigrationTable(db); err != nil {
		return err
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	names, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}

	for _, name := range names {
		if _, exists := applied[name]; exists {
			continue
		}

		if err := applyMigration(db, name); err != nil {
			return err
		}
	}

	return nil
}

func createMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
  		CREATE TABLE IF NOT EXISTS schema_migrations (
  			version TEXT PRIMARY KEY,
  			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
  		);
  	`)

	if err != nil {
		return fmt.Errorf("create migration history table: %w", err)
	}

	return nil
}

func getAppliedMigrations(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query migration history: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("read migration version: %w", err)
		}
		applied[version] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration history: %w", err)
	}

	return applied, nil
}

func applyMigration(db *sql.DB, name string) error {
	script, err := files.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(script)); err != nil {
		return fmt.Errorf("execute migration %s: %w", name, err)
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
