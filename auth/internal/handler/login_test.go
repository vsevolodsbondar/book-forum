package handler

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"uuid"

	"auth/internal/database"
	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/service"
)

const loginTestPassword = "correct horse battery staple"

func TestLogin(t *testing.T) {
	handler, db := newTestLoginHandler(t)

	recorder := sendLogin(handler, loginBody(t, " User@Example.com ", loginTestPassword))
	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var data LoginResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	if data.User.ID <= 0 {
		t.Fatalf("invalid user ID: %d", data.User.ID)
	}
	if data.User.Email != "user@example.com" {
		t.Fatalf("got email %q, want user@example.com", data.User.Email)
	}
	if data.User.Username != "User" {
		t.Fatalf("got username %q, want User", data.User.Username)
	}
	if data.Session.ID == "" || data.Session.Token == "" {
		t.Fatalf("invalid session response: %+v", data.Session)
	}
	if data.Session.ExpiresAt <= time.Now().Unix() {
		t.Fatalf("session already expired: %d", data.Session.ExpiresAt)
	}

	sessionID, err := uuid.Parse(data.Session.ID)
	if err != nil {
		t.Fatalf("parse session ID: %v", err)
	}
	if sessionID[6]>>4 != 4 {
		t.Fatalf("got UUID version %d, want 4", sessionID[6]>>4)
	}

	secret, err := base64.RawURLEncoding.Strict().DecodeString(data.Session.Token)
	if err != nil {
		t.Fatalf("decode session token: %v", err)
	}
	if len(secret) != 32 {
		t.Fatalf("got token length %d bytes, want 32", len(secret))
	}

	var (
		storedID   string
		userID     int64
		storedHash []byte
		createdAt  int64
		lastSeenAt int64
		expiresAt  int64
	)

	err = db.QueryRow("SELECT id, user_id, token_hash, created_at, last_seen_at, expires_at FROM sessions").Scan(
		&storedID,
		&userID,
		&storedHash,
		&createdAt,
		&lastSeenAt,
		&expiresAt,
	)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}

	expectedHash := sha256.Sum256(secret)

	if storedID != data.Session.ID {
		t.Fatalf("got stored session ID %q, want %q", storedID, data.Session.ID)
	}
	if userID != data.User.ID {
		t.Fatalf("got stored user ID %d, want %d", userID, data.User.ID)
	}
	if !bytes.Equal(storedHash, expectedHash[:]) {
		t.Fatal("stored token hash does not match session token")
	}
	if bytes.Equal(storedHash, secret) || bytes.Equal(storedHash, []byte(data.Session.Token)) {
		t.Fatal("raw or encoded session token was persisted")
	}
	if createdAt != lastSeenAt {
		t.Fatalf("created_at %d differs from last_seen_at %d", createdAt, lastSeenAt)
	}
	if expiresAt != data.Session.ExpiresAt {
		t.Fatalf("got stored expiry %d, want %d", expiresAt, data.Session.ExpiresAt)
	}
	if expiresAt-createdAt != int64(time.Hour/time.Second) {
		t.Fatalf("got session lifetime %d seconds, want 3600", expiresAt-createdAt)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if len(envelope) != 2 {
		t.Fatalf("unexpected response fields: %s", recorder.Body.String())
	}

	var userFields map[string]json.RawMessage
	if err := json.Unmarshal(envelope["user"], &userFields); err != nil {
		t.Fatalf("decode user fields: %v", err)
	}
	for _, field := range []string{"id", "email", "username"} {
		if _, ok := userFields[field]; !ok {
			t.Fatalf("missing user field %q", field)
		}
	}
	if len(userFields) != 3 {
		t.Fatalf("unexpected user fields: %s", envelope["user"])
	}

	var sessionFields map[string]json.RawMessage
	if err := json.Unmarshal(envelope["session"], &sessionFields); err != nil {
		t.Fatalf("decode session fields: %v", err)
	}
	for _, field := range []string{"id", "token", "expires_at"} {
		if _, ok := sessionFields[field]; !ok {
			t.Fatalf("missing session field %q", field)
		}
	}
	if len(sessionFields) != 3 {
		t.Fatalf("unexpected session fields: %s", envelope["session"])
	}

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessions != 1 {
		t.Fatalf("got %d sessions, want 1", sessions)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	handler, db := newTestLoginHandler(t)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"unknown_email", "unknown@example.com", loginTestPassword},
		{"incorrect_password", "user@example.com", "incorrect password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := sendLogin(handler, loginBody(t, tt.email, tt.password))
			assertErrorResponse(t, recorder, http.StatusUnauthorized, "invalid_credentials")
		})
	}

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessions != 0 {
		t.Fatalf("got %d sessions, want 0", sessions)
	}
}

func TestLoginInvalidRequest(t *testing.T) {
	handler, db := newTestLoginHandler(t)

	tests := []struct {
		name string
		body string
	}{
		{"empty", ""},
		{"malformed", "{"},
		{"null", "null"},
		{"array", "[]"},
		{"wrong_type", `{"email":42}`},
		{"multiple_values", `{} {}`},
		{"missing_fields", `{}`},
		{"invalid_email", loginBody(t, "invalid", loginTestPassword)},
		{"unicode_email", loginBody(t, "üser@example.com", loginTestPassword)},
		{"missing_password", `{"email":"user@example.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := sendLogin(handler, tt.body)
			assertErrorResponse(t, recorder, http.StatusBadRequest, "invalid_request")
		})
	}

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessions != 0 {
		t.Fatalf("invalid requests created %d sessions", sessions)
	}
}

func TestLoginBodyLimit(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		status int
	}{
		{"at_limit", 8192, http.StatusOK},
		{"over_limit", 8193, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestLoginHandler(t)
			body := loginBody(t, "user@example.com", loginTestPassword)
			body += strings.Repeat(" ", tt.size-len(body))

			recorder := sendLogin(handler, body)
			if tt.status == http.StatusRequestEntityTooLarge {
				assertErrorResponse(t, recorder, tt.status, "request_too_large")
				return
			}

			if recorder.Code != tt.status {
				t.Fatalf("got status %d, want %d: %s", recorder.Code, tt.status, recorder.Body.String())
			}
		})
	}
}

func TestLoginDatabaseFailure(t *testing.T) {
	handler, db := newTestLoginHandler(t)

	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	recorder := sendLogin(handler, loginBody(t, "user@example.com", loginTestPassword))
	assertErrorResponse(t, recorder, http.StatusInternalServerError, "internal_error")
}

func TestLoginInvalidStoredHash(t *testing.T) {
	handler, db := newTestLoginHandler(t)

	if _, err := db.Exec("UPDATE users SET password_hash = ?", "invalid hash"); err != nil {
		t.Fatalf("corrupt password hash: %v", err)
	}

	recorder := sendLogin(handler, loginBody(t, "user@example.com", loginTestPassword))
	assertErrorResponse(t, recorder, http.StatusInternalServerError, "internal_error")

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessions != 0 {
		t.Fatalf("invalid hash created %d sessions", sessions)
	}
}

func newTestLoginHandler(t *testing.T) (http.HandlerFunc, *sql.DB) {
	t.Helper()

	sessionHandler, db := newTestSessionHandler(t)
	return sessionHandler.Login, db
}

func newTestSessionHandler(t *testing.T) (*SessionHandler, *sql.DB) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	hasher, err := password.New(19456, 2, 1, 2)
	if err != nil {
		t.Fatalf("create password hasher: %v", err)
	}

	validate, err := NewValidator()
	if err != nil {
		t.Fatalf("create request validator: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, hasher)

	if _, err := userService.Register(t.Context(), "user@example.com", "User", loginTestPassword); err != nil {
		t.Fatalf("register user: %v", err)
	}

	sessionRepo := repository.NewSessionRepository(db)
	sessionService, err := service.NewSessionService(userRepo, sessionRepo, hasher, time.Hour, 30*time.Minute)
	if err != nil {
		t.Fatalf("create session service: %v", err)
	}

	sessionHandler := NewSessionHandler(sessionService, validate)

	return sessionHandler, db
}

func loginBody(t *testing.T, email, password string) string {
	t.Helper()

	body, err := json.Marshal(LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Fatalf("encode login request: %v", err)
	}

	return string(body)
}

func sendLogin(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/v1/login", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	handler(recorder, request)

	return recorder
}
