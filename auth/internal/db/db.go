package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Open prepares the database directory and returns a verified SQLite connection pool.
func Open(path string) (*sql.DB, error) {
	// Creates the database's parent directory and any missing ancestors.
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Sets up a database connection pool.
	database, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verifies connectivity, opening a connection if needed.
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("verify database connection: %w", err)
	}

	return database, nil
}
