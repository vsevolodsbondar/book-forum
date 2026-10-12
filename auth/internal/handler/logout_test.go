package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogout(t *testing.T) {
	handler, db := newTestSessionHandler(t)
	login := sendLogin(handler.Login, loginBody(t, "user@example.com", loginTestPassword))

	var data LoginResponse
	if err := json.Unmarshal(login.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	recorder := sendLogout(handler.Logout, "Bearer "+data.Session.Token)
	assertNoContent(t, recorder)

	var sessions int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessions); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessions != 0 {
		t.Fatalf("got %d sessions, want 0", sessions)
	}

	recorder = sendValidation(handler.Validate, "Bearer "+data.Session.Token)
	assertErrorResponse(t, recorder, http.StatusUnauthorized, "invalid_session")

	recorder = sendLogout(handler.Logout, "Bearer "+data.Session.Token)
	assertNoContent(t, recorder)
}

func TestLogoutTokenOutcomes(t *testing.T) {
	handler, _ := newTestSessionHandler(t)
	unknown := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))

	tests := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"missing_token", "Bearer"},
		{"wrong_scheme", "Basic token"},
		{"extra_value", "Bearer token extra"},
		{"malformed", "Bearer !!!"},
		{"padded", "Bearer YWJjZA=="},
		{"short", "Bearer YWJjZA"},
		{"unknown", "Bearer " + unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertNoContent(t, sendLogout(handler.Logout, tt.header))
		})
	}
}

func TestLogoutDatabaseFailure(t *testing.T) {
	handler, db := newTestSessionHandler(t)
	token := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))

	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	recorder := sendLogout(handler.Logout, "Bearer "+token)
	assertErrorResponse(t, recorder, http.StatusInternalServerError, "internal_error")

	assertNoContent(t, sendLogout(handler.Logout, "Bearer malformed"))
}

func sendLogout(handler http.HandlerFunc, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/v1/logout", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertNoContent(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("got response body %q, want empty body", recorder.Body.String())
	}
}
