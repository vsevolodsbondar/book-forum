package client

import (
	"context"
	"encoding/json"
	"errors"
	e "forum_backend/custom_err"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHTTPClient_RegisterUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/v1/register" {
			t.Errorf("expected path /v1/register, got %s", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	result, err := client.RegisterUser(
		context.Background(),
		RegisterUserRequestDTO{},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != (RegisterUserResponseDTO{}) {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestAuthHTTPClient_LoginUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/v1/login" {
			t.Errorf("expected path /v1/login, got %s", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	result, err := client.LoginUser(
		context.Background(),
		LoginUserRequestDTO{},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != (LoginUserResponseDTO{}) {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestAuthHTTPClient_ValidateSession(t *testing.T) {
	const token = "test-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/v1/session/validate" {
			t.Errorf("expected path /v1/session/validate, got %s", r.URL.Path)
		}

		expectedAuth := "Bearer " + token

		if got := r.Header.Get("Authorization"); got != expectedAuth {
			t.Errorf(
				"expected Authorization %q, got %q",
				expectedAuth,
				got,
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	result, err := client.ValidateSession(
		context.Background(),
		token,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != (ValidateSessionResponseDTO{}) {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestAuthHTTPClient_LogoutUser(t *testing.T) {
	const token = "test-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/v1/logout" {
			t.Errorf("expected path /v1/logout, got %s", r.URL.Path)
		}

		expectedAuth := "Bearer " + token

		if got := r.Header.Get("Authorization"); got != expectedAuth {
			t.Errorf(
				"expected Authorization %q, got %q",
				expectedAuth,
				got,
			)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	err := client.LogoutUser(
		context.Background(),
		token,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAuthHTTPClient_AuthServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)

		_, _ = w.Write([]byte(`{
			"message": "user already exists"
		}`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := client.RegisterUser(
		context.Background(),
		RegisterUserRequestDTO{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, e.ErrAuthService) {
		t.Errorf(
			"expected ErrAuthService, got %v",
			err,
		)
	}

	if !strings.Contains(err.Error(), "user already exists") {
		t.Errorf(
			"expected error to contain response message, got %v",
			err,
		)
	}
}

func TestAuthHTTPClient_InvalidErrorResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := client.RegisterUser(
		context.Background(),
		RegisterUserRequestDTO{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, e.ErrJSONDecodeFailed) {
		t.Errorf(
			"expected ErrJSONDecodeFailed, got %v",
			err,
		)
	}
}

func TestAuthHTTPClient_InvalidSuccessResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	client := &AuthHTTPClient{
		BaseURL: server.URL,
		Client:  server.Client(),
	}

	_, err := client.RegisterUser(
		context.Background(),
		RegisterUserRequestDTO{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, e.ErrJSONDecodeFailed) {
		t.Errorf(
			"expected ErrJSONDecodeFailed, got %v",
			err,
		)
	}
}
