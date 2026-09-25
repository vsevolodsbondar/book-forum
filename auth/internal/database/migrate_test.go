package database

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	db := newTestDB(t)
	var version string

	err := db.QueryRow(
		"SELECT version FROM schema_migrations",
	).Scan(&version)
	if err != nil {
		t.Fatalf("query migration history: %v", err)
	}

	if version != "migrations/0001_init.sql" {
		t.Fatalf("unexpected migration version: %q", version)
	}

	assertTablesExist(t, db, "users", "sessions", "schema_migrations")
}

func TestMigrateAgain(t *testing.T) {
	db := newTestDB(t)

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate again: %v", err)
	}

	var count int
	if err := db.QueryRow(
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

	db, err := Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	return db
}

func assertTablesExist(t *testing.T, db *sql.DB, names ...string) {
	t.Helper()

	for _, name := range names {
		var count int
		err := db.QueryRow(
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
