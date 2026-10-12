package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"auth/internal/database"
	"auth/internal/repository"
	"auth/internal/service"
)

func TestHealthLive(t *testing.T) {
	health, db := newTestHealthHandler(t, true)

	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	recorder := httptest.NewRecorder()

	health.Live(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("got content type %q, want application/json",
			recorder.Header().Get("Content-Type"))
	}

	var body HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Service != "auth" {
		t.Fatalf("got service %q, want auth", body.Service)
	}
	if body.Status != "ok" {
		t.Fatalf("got status %q, want ok", body.Status)
	}
	if body.Checks != nil {
		t.Fatalf("got checks %v, want checks omitted", body.Checks)
	}

	if strings.Contains(recorder.Body.String(), `"checks"`) {
		t.Fatalf("liveness response should omit checks: %s", recorder.Body.String())
	}
}

func TestHealthReady(t *testing.T) {
	tests := []struct {
		name          string
		migrate       bool
		insertUser    bool
		closeDatabase bool
		cancelContext bool
		wantHTTP      int
		wantStatus    string
		wantDBStatus  string
	}{
		{
			name:         "empty_database",
			migrate:      true,
			wantHTTP:     http.StatusOK,
			wantStatus:   "ok",
			wantDBStatus: "ok",
		},
		{
			name:         "populated_database",
			migrate:      true,
			insertUser:   true,
			wantHTTP:     http.StatusOK,
			wantStatus:   "ok",
			wantDBStatus: "ok",
		},
		{
			name:         "users_table_missing",
			migrate:      false,
			wantHTTP:     http.StatusServiceUnavailable,
			wantStatus:   "unavailable",
			wantDBStatus: "unavailable",
		},
		{
			name:          "database_closed",
			migrate:       true,
			closeDatabase: true,
			wantHTTP:      http.StatusServiceUnavailable,
			wantStatus:    "unavailable",
			wantDBStatus:  "unavailable",
		},
		{
			name:          "request_context_cancelled",
			migrate:       true,
			cancelContext: true,
			wantHTTP:      http.StatusServiceUnavailable,
			wantStatus:    "unavailable",
			wantDBStatus:  "unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health, db := newTestHealthHandler(t, tt.migrate)

			if tt.insertUser {
				if _, err := db.Exec(
					"INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)",
					"user@example.com", "User", "hash",
				); err != nil {
					t.Fatalf("insert user: %v", err)
				}
			}

			if tt.closeDatabase {
				if err := db.Close(); err != nil {
					t.Fatalf("close database: %v", err)
				}
			}

			request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

			if tt.cancelContext {
				ctx, cancel := context.WithCancel(request.Context())
				cancel()
				request = request.WithContext(ctx)
			}

			recorder := httptest.NewRecorder()
			health.Ready(recorder, request)

			if recorder.Code != tt.wantHTTP {
				t.Fatalf("got status %d, want %d: %s",
					recorder.Code, tt.wantHTTP, recorder.Body.String())
			}
			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("got content type %q, want application/json",
					recorder.Header().Get("Content-Type"))
			}

			var body HealthResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if body.Service != "auth" {
				t.Fatalf("got service %q, want auth", body.Service)
			}
			if body.Status != tt.wantStatus {
				t.Fatalf("got status %q, want %q", body.Status, tt.wantStatus)
			}

			databaseCheck, ok := body.Checks["database"]
			if !ok {
				t.Fatalf("missing database check: %s", recorder.Body.String())
			}
			if databaseCheck.Status != tt.wantDBStatus {
				t.Fatalf("got database status %q, want %q",
					databaseCheck.Status, tt.wantDBStatus)
			}

			if strings.Contains(recorder.Body.String(), "no such table") {
				t.Fatalf("response exposed database error: %s", recorder.Body.String())
			}
		})
	}
}

func newTestHealthHandler(t *testing.T, migrate bool) (*HealthHandler, *sql.DB) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	if migrate {
		if err := database.Migrate(db); err != nil {
			t.Fatalf("migrate database: %v", err)
		}
	}

	repo := repository.NewHealthRepository(db)
	healthService := service.NewHealthService(repo)

	return NewHealthHandler(healthService), db
}
