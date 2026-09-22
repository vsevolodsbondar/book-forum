package handler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/service"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{"invalid_request", ErrInvalidRequest, http.StatusBadRequest, "invalid_request", "invalid request data"},
		{"invalid_credentials", service.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials", "invalid email or password"},
		{"invalid_session", service.ErrInvalidSession, http.StatusUnauthorized, "invalid_session", "invalid or expired session"},
		{"identity_conflict", repository.ErrIdentityConflict, http.StatusConflict, "identity_conflict", "email or username already in use"},
		{"busy", password.ErrBusy, http.StatusServiceUnavailable, "service_busy", "service temporarily busy"},
		{"body_limit", &http.MaxBytesError{Limit: 8192}, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds 8192 bytes"},
		{"unexpected", errors.New("private database details"), http.StatusInternalServerError, "internal_error", "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variants := []struct {
				name string
				err  error
			}{
				{"direct", tt.err},
				{"wrapped", fmt.Errorf("operation failed: %w", tt.err)},
			}

			for _, variant := range variants {
				t.Run(variant.name, func(t *testing.T) {
					status, data := mapError(variant.err)

					if status != tt.status {
						t.Fatalf("got status %d, want %d", status, tt.status)
					}
					if data.Error.Code != tt.code {
						t.Fatalf("got code %q, want %q", data.Error.Code, tt.code)
					}
					if data.Error.Message != tt.message {
						t.Fatalf("got message %q, want %q", data.Error.Message, tt.message)
					}
				})
			}
		})
	}
}
