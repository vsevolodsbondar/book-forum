package db

import (
	"database/sql"
	"fmt"
	"forum_backend/migrations"

	_ "github.com/mattn/go-sqlite3"
)

func Init() (*sql.DB, error) {
	data, err := sql.Open("sqlite3", "./forum.db?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000")
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
	//we need to decide what and how gonna print or log messages
	fmt.Println("Connected to SQLite")
	return data, nil
}
