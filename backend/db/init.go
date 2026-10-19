package db

import (
	"database/sql"
	"fmt"
	"forum_backend/migrations"

	_ "github.com/mattn/go-sqlite3"
)

func Init(isSeeding bool) (*sql.DB, error) {
	data, err := sql.Open("sqlite3", "./db/forum.db?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	//check if no errors with connection
	if err := data.Ping(); err != nil {
		data.Close()
		return nil, err
	}
	if err := migrations.Run(data); err != nil {
		data.Close()
		return nil, err
	}

	if isSeeding {
		seeded, err := isDatabaseSeeded(data)
		if err != nil {
			data.Close()
			return nil, err
		}

		if !seeded {
			if err := seedDB(data); err != nil {
				data.Close()
				return nil, err
			}
		} else {
			fmt.Println("Seed data already exists, skipping seeding")
		}
	}

	fmt.Println("Connected to SQLite")
	return data, nil
}

func isDatabaseSeeded(db *sql.DB) (bool, error) {
	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check seed data: %w", err)
	}

	return count > 0, nil
}
