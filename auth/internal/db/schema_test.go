package db

import "testing"

func TestUniqueEmail(t *testing.T) {
	database := newTestDB(t)

	_, err := database.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		"user@example.com", "first", "test-hash",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = database.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		"user@example.com", "second", "test-hash",
	)
	if err == nil {
		t.Fatal("expected duplicate email to be rejected")
	}
}

func TestUniqueUsername(t *testing.T) {
	database := newTestDB(t)

	_, err := database.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		"first@example.com", "User", "test-hash",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = database.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		"second@example.com", "user", "test-hash",
	)
	if err == nil {
		t.Fatal("expected case-insensitive duplicate username to be rejected")
	}
}

func TestIDNotReused(t *testing.T) {
	database := newTestDB(t)
	query := "INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)"

	result, err := database.Exec(query, "first@example.com", "first", "test-hash")
	if err != nil {
		t.Fatalf("insert first user: %v", err)
	}
	firstID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read first user ID: %v", err)
	}

	if _, err := database.Exec("DELETE FROM users WHERE id = ?", firstID); err != nil {
		t.Fatalf("delete first user: %v", err)
	}

	result, err = database.Exec(query, "second@example.com", "second", "test-hash")
	if err != nil {
		t.Fatalf("insert second user: %v", err)
	}
	secondID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read second user ID: %v", err)
	}

	if secondID <= firstID {
		t.Fatalf("got user ID %d, want greater than %d", secondID, firstID)
	}
}

func TestSessionCascadeDelete(t *testing.T) {
	database := newTestDB(t)

	result, err := database.Exec(
		"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
		"user@example.com", "User", "test-hash",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read user ID: %v", err)
	}

	_, err = database.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, unixepoch() + 3600)
	`, "test-session", userID, make([]byte, 32))
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	if _, err := database.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var count int
	if err := database.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE user_id = ?", userID,
	).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}

	if count != 0 {
		t.Fatalf("got %d sessions after deleting user, want 0", count)
	}
}

func TestSessionIndexes(t *testing.T) {
	database := newTestDB(t)

	for _, name := range []string{
		"idx_sessions_user_id",
		"idx_sessions_expires_at",
	} {
		var count int
		err := database.QueryRow(
			`SELECT COUNT(*) FROM sqlite_master
			 WHERE type = 'index' AND tbl_name = 'sessions' AND name = ?`,
			name,
		).Scan(&count)
		if err != nil {
			t.Fatalf("check index %s: %v", name, err)
		}
		if count != 1 {
			t.Fatalf("expected index %s to exist", name)
		}
	}
}