package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"auth/internal/database"
	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/response"
	"auth/internal/service"
)

const registerTestPassword = " correct horse battery staple "

func TestRegister(t *testing.T) {
	handler, db, hasher := newTestRegisterHandler(t)
	body := registrationBody(t, " User@Example.com ", " User ", registerTestPassword)

	recorder := sendRegistration(handler, body)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201: %s", recorder.Code, recorder.Body.String())
	}

	var data RegisterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode registration response: %v", err)
	}

	if data.User.ID <= 0 || data.User.CreatedAt <= 0 {
		t.Fatalf("invalid user identity: %+v", data.User)
	}
	if data.User.Email != "user@example.com" {
		t.Fatalf("got email %q, want user@example.com", data.User.Email)
	}
	if data.User.Username != "User" {
		t.Fatalf("got username %q, want User", data.User.Username)
	}

	var storedHash string
	if err := db.QueryRow("SELECT password_hash FROM users WHERE id = ?", data.User.ID).Scan(&storedHash); err != nil {
		t.Fatalf("query stored password hash: %v", err)
	}

	match, err := hasher.Verify(registerTestPassword, storedHash)
	if err != nil || !match {
		t.Fatalf("verify stored password: match=%v, err=%v", match, err)
	}

	match, err = hasher.Verify(strings.TrimSpace(registerTestPassword), storedHash)
	if err != nil || match {
		t.Fatalf("verify trimmed password: match=%v, err=%v", match, err)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if len(envelope) != 1 {
		t.Fatalf("unexpected response fields: %s", recorder.Body.String())
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(envelope["user"], &fields); err != nil {
		t.Fatalf("decode user fields: %v", err)
	}
	for _, field := range []string{"id", "email", "username", "created_at"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("missing user field %q", field)
		}
	}
	if len(fields) != 4 {
		t.Fatalf("unexpected user fields: %s", envelope["user"])
	}

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("query sessions: %v", err)
	}
	if sessions != 0 {
		t.Fatalf("got %d sessions, want 0", sessions)
	}
}

func TestRegisterIdentityConflict(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		username string
	}{
		{"email", " USER@EXAMPLE.COM ", "Other"},
		{"username", "other@example.com", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, db, _ := newTestRegisterHandler(t)

			first := sendRegistration(handler, registrationBody(t, "user@example.com", "User", registerTestPassword))
			if first.Code != http.StatusCreated {
				t.Fatalf("initial registration: %d: %s", first.Code, first.Body.String())
			}

			recorder := sendRegistration(handler, registrationBody(t, tt.email, tt.username, registerTestPassword))
			assertRegistrationError(t, recorder, http.StatusConflict, "identity_conflict")

			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
				t.Fatalf("query users: %v", err)
			}
			if count != 1 {
				t.Fatalf("got %d users, want 1", count)
			}
		})
	}
}

func TestRegisterConcurrentIdentityConflict(t *testing.T) {
	handler, db, _ := newTestRegisterHandler(t)
	body := registrationBody(t, "user@example.com", "User", registerTestPassword)

	start := make(chan struct{})
	statuses := make(chan int, 2)
	var workers sync.WaitGroup

	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			statuses <- sendRegistration(handler, body).Code
		}()
	}

	close(start)
	workers.Wait()
	close(statuses)

	counts := make(map[int]int)
	for status := range statuses {
		counts[status]++
	}

	if counts[http.StatusCreated] != 1 || counts[http.StatusConflict] != 1 {
		t.Fatalf("got statuses %v, want one 201 and one 409", counts)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("query users: %v", err)
	}
	if count != 1 {
		t.Fatalf("got %d users, want 1", count)
	}
}

func TestRegisterInvalidRequest(t *testing.T) {
	handler, db, _ := newTestRegisterHandler(t)

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
		{"invalid_email", registrationBody(t, "invalid", "User", registerTestPassword)},
		{"unicode_email", registrationBody(t, "üser@example.com", "User", registerTestPassword)},
		{"quoted_email", registrationBody(t, `"user"@example.com`, "User", registerTestPassword)},
		{"domain_literal", registrationBody(t, "user@[127.0.0.1]", "User", registerTestPassword)},
		{"domain_without_dot", registrationBody(t, "user@example", "User", registerTestPassword)},
		{"long_local_part", registrationBody(t, strings.Repeat("a", 65)+"@example.com", "User", registerTestPassword)},
		{"long_email", registrationBody(t, strings.Repeat("a", 64)+"@"+strings.Repeat("b", 63)+"."+strings.Repeat("c", 63)+"."+strings.Repeat("d", 62), "User", registerTestPassword)},
		{"short_username", registrationBody(t, "user@example.com", "Us", registerTestPassword)},
		{"long_username", registrationBody(t, "user@example.com", strings.Repeat("a", 31), registerTestPassword)},
		{"unicode_username", registrationBody(t, "user@example.com", "Üser", registerTestPassword)},
		{"username_symbol", registrationBody(t, "user@example.com", "User!", registerTestPassword)},
		{"short_password", registrationBody(t, "user@example.com", "User", strings.Repeat("a", 14))},
		{"long_password", registrationBody(t, "user@example.com", "User", strings.Repeat("a", 129))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := sendRegistration(handler, tt.body)
			assertRegistrationError(t, recorder, http.StatusBadRequest, "invalid_request")
		})
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("query users: %v", err)
	}
	if count != 0 {
		t.Fatalf("invalid requests created %d users", count)
	}
}

func TestRegisterPasswordLength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		status   int
	}{
		{"minimum", strings.Repeat("a", 15), http.StatusCreated},
		{"maximum", strings.Repeat("a", 128), http.StatusCreated},
		{"unicode_minimum", strings.Repeat("界", 15), http.StatusCreated},
		{"unicode_too_short", strings.Repeat("界", 14), http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _ := newTestRegisterHandler(t)
			recorder := sendRegistration(handler, registrationBody(t, "user@example.com", "User", tt.password))

			if recorder.Code != tt.status {
				t.Fatalf("got status %d, want %d: %s", recorder.Code, tt.status, recorder.Body.String())
			}
		})
	}
}

func TestRegisterBodyLimit(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		status int
	}{
		{"at_limit", 8192, http.StatusCreated},
		{"over_limit", 8193, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _ := newTestRegisterHandler(t)
			body := registrationBody(t, "user@example.com", "User", registerTestPassword)
			body += strings.Repeat(" ", tt.size-len(body))

			recorder := sendRegistration(handler, body)
			if tt.status == http.StatusRequestEntityTooLarge {
				assertRegistrationError(t, recorder, tt.status, "request_too_large")
				return
			}

			if recorder.Code != tt.status {
				t.Fatalf("got status %d, want %d: %s",
					recorder.Code, tt.status, recorder.Body.String())
			}
		})
	}
}

func TestRegisterDatabaseFailure(t *testing.T) {
	handler, db, _ := newTestRegisterHandler(t)
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	recorder := sendRegistration(handler, registrationBody(t, "user@example.com", "User", registerTestPassword))
	assertRegistrationError(t, recorder, http.StatusInternalServerError, "internal_error")

	var data response.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if data.Error.Message != "internal server error" {
		t.Fatalf("unexpected public message: %q", data.Error.Message)
	}
}

func newTestRegisterHandler(t *testing.T) (http.HandlerFunc, *sql.DB, *password.Hasher) {
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

	return Register(userService, validate), db, hasher
}

func registrationBody(t *testing.T, email, username, password string) string {
	t.Helper()

	body, err := json.Marshal(RegisterRequest{
		Email:    email,
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("encode registration request: %v", err)
	}

	return string(body)
}

func sendRegistration(handler http.Handler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/v1/register", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = -1

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder
}

func assertRegistrationError(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()

	if recorder.Code != status {
		t.Fatalf("got status %d, want %d: %s", recorder.Code, status, recorder.Body.String())
	}

	var data response.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if data.Error.Code != code {
		t.Fatalf("got error code %q, want %q", data.Error.Code, code)
	}
	if data.Error.Message == "" {
		t.Fatal("missing error message")
	}
}
