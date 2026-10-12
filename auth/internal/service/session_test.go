package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"auth/internal/database"
	"auth/internal/repository"
)

func TestValidateSessionBoundaries(t *testing.T) {
	const now int64 = 2_000_000

	tests := []struct {
		name       string
		lastSeenAt int64
		expiresAt  int64
		valid      bool
	}{
		{"before_absolute_expiry", now - 10, now + 1, true},
		{"at_absolute_expiry", now - 10, now, false},
		{"before_idle_expiry", now - 59, now + 100, true},
		{"at_idle_expiry", now - 60, now + 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, db, token := newValidationTestService(t, now, tt.lastSeenAt, tt.expiresAt)

			result, err := service.Validate(t.Context(), token)
			if tt.valid {
				if err != nil {
					t.Fatalf("validate session: %v", err)
				}
				if result.UserID != 1 || result.Username != "User" || result.SessionID != "session-id" || result.ExpiresAt != tt.expiresAt {
					t.Fatalf("unexpected validation result: %+v", result)
				}
			} else if !errors.Is(err, ErrInvalidSession) {
				t.Fatalf("got error %v, want ErrInvalidSession", err)
			}

			var lastSeenAt, expiresAt int64
			if err := db.QueryRow("SELECT last_seen_at, expires_at FROM sessions WHERE id = ?", "session-id").Scan(&lastSeenAt, &expiresAt); err != nil {
				t.Fatalf("query session: %v", err)
			}
			if tt.valid && lastSeenAt != now {
				t.Fatalf("got last_seen_at %d, want %d", lastSeenAt, now)
			}
			if !tt.valid && lastSeenAt != tt.lastSeenAt {
				t.Fatalf("expired session activity changed from %d to %d", tt.lastSeenAt, lastSeenAt)
			}
			if expiresAt != tt.expiresAt {
				t.Fatalf("expiry changed from %d to %d", tt.expiresAt, expiresAt)
			}

			if !tt.valid {
				service.now = func() time.Time { return time.Unix(now+1, 0) }
				if _, err := service.Validate(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
					t.Fatalf("expired session was revived: %v", err)
				}
			}
		})
	}
}

func TestValidateSessionUnknownToken(t *testing.T) {
	service, _, _ := newValidationTestService(t, 2_000_000, 1_999_990, 2_000_100)
	unknown := base64.RawURLEncoding.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))

	if _, err := service.Validate(t.Context(), unknown); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("got error %v, want ErrInvalidSession", err)
	}
}

func TestLogoutDeletesExpiredSession(t *testing.T) {
	tests := []struct {
		name       string
		lastSeenAt int64
		expiresAt  int64
	}{
		{"absolute_expiry", 1_999_990, 2_000_000},
		{"idle_expiry", 1_999_940, 2_000_100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, db, token := newValidationTestService(t, 2_000_000, tt.lastSeenAt, tt.expiresAt)

			if err := service.Logout(t.Context(), token); err != nil {
				t.Fatalf("logout: %v", err)
			}

			var sessions int
			if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
				t.Fatalf("count sessions: %v", err)
			}
			if sessions != 0 {
				t.Fatalf("got %d sessions, want 0", sessions)
			}
		})
	}
}

func TestLogoutInvalidTokenSkipsDatabase(t *testing.T) {
	service, db, _ := newValidationTestService(t, 2_000_000, 1_999_990, 2_000_100)
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	for _, token := range []string{"malformed", "YWJjZA==", "YWJjZA"} {
		if err := service.Logout(t.Context(), token); err != nil {
			t.Fatalf("logout token %q: %v", token, err)
		}
	}
}

func newValidationTestService(t *testing.T, now, lastSeenAt, expiresAt int64) (*SessionService, *sql.DB, string) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	if _, err := db.Exec("INSERT INTO users (id, email, username, password_hash) VALUES (?, ?, ?, ?)", 1, "user@example.com", "User", "hash"); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	secret := []byte("0123456789abcdefghijklmnopqrstuv")
	tokenHash := sha256.Sum256(secret)
	if _, err := db.Exec(`
		INSERT INTO sessions (id, user_id, token_hash, created_at, last_seen_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "session-id", 1, tokenHash[:], now-1000, lastSeenAt, expiresAt); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	service := &SessionService{
		sessionRepo:        repository.NewSessionRepository(db),
		sessionIdleTimeout: time.Minute,
		now:                func() time.Time { return time.Unix(now, 0) },
	}

	return service, db, base64.RawURLEncoding.EncodeToString(secret)
}
