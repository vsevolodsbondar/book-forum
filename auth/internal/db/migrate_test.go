package db

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	database := newTestDB(t)
	var version string

	err := database.QueryRow(
		"SELECT version FROM schema_migrations",
	).Scan(&version)
	if err != nil {
		t.Fatalf("query migration history: %v", err)
	}

	if version != "migrations/0001_init.sql" {
		t.Fatalf("unexpected migration version: %q", version)
	}

	assertTablesExist(t, database, "users", "sessions", "schema_migrations")
}


func TestMigrateAgain(t *testing.T) {
  	database := newTestDB(t)

  	if err := Migrate(database); err != nil {
  		t.Fatalf("migrate again: %v", err)
  	}

  	var count int
  	if err := database.QueryRow(
  		"SELECT COUNT(*) FROM schema_migrations",
  	).Scan(&count); err != nil {
  		t.Fatalf("count migration history: %v", err)
  	}

  	if count != 1 {
  		t.Fatalf("got %d migrations, want 1", count)
  	}
  }

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
	})

	if err := Migrate(database); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	return database
}

func assertTablesExist(t *testing.T, database *sql.DB, names ...string) {
	t.Helper()

	for _, name := range names {
		var count int
		err := database.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
			name,
		).Scan(&count)
		if err != nil {
			t.Fatalf("check table %s: %v", name, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist", name)
		}
	}
}
