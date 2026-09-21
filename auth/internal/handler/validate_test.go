package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateSession(t *testing.T) {
	handler, _ := newTestSessionHandler(t)
	login := sendLogin(handler.Login, loginBody(t, "user@example.com", loginTestPassword))

	var loginData LoginResponse
	if err := json.Unmarshal(login.Body.Bytes(), &loginData); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	recorder := sendValidation(handler.Validate, "Bearer "+loginData.Session.Token)
	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var data ValidationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}
	if data.User.ID != loginData.User.ID || data.User.Username != loginData.User.Username {
		t.Fatalf("got user %+v, want ID %d and username %q", data.User, loginData.User.ID, loginData.User.Username)
	}
	if data.Session.ID != loginData.Session.ID || data.Session.ExpiresAt != loginData.Session.ExpiresAt {
		t.Fatalf("got session %+v, want ID %q and expiry %d", data.Session, loginData.Session.ID, loginData.Session.ExpiresAt)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if len(envelope) != 2 {
		t.Fatalf("unexpected response fields: %s", recorder.Body.String())
	}
}

func TestValidateSessionInvalidToken(t *testing.T) {
	handler, _ := newTestSessionHandler(t)
	randomToken := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))

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
		{"unknown", "Bearer " + randomToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := sendValidation(handler.Validate, tt.header)
			assertErrorResponse(t, recorder, http.StatusUnauthorized, "invalid_session")
		})
	}
}

func TestValidateSessionDatabaseFailure(t *testing.T) {
	handler, db := newTestSessionHandler(t)
	token := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("x", 32)))

	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	recorder := sendValidation(handler.Validate, "Bearer "+token)
	assertErrorResponse(t, recorder, http.StatusInternalServerError, "internal_error")
}

func sendValidation(handler http.HandlerFunc, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/v1/session/validate", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
