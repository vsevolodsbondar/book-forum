package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/response"
	"auth/internal/service"
)

// ErrInvalidRequest is returned when request data is malformed or invalid.
var ErrInvalidRequest = errors.New("invalid request")

func mapError(err error) (int, response.ErrorResponse) {
	var sizeErr *http.MaxBytesError

	switch {
	case errors.As(err, &sizeErr):
		return http.StatusRequestEntityTooLarge, response.NewError("request_too_large", "request body exceeds 8192 bytes")
	case errors.Is(err, ErrInvalidRequest):
		return http.StatusBadRequest, response.NewError("invalid_request", "invalid request data")
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, response.NewError("invalid_credentials", "invalid email or password")
	case errors.Is(err, repository.ErrIdentityConflict):
		return http.StatusConflict, response.NewError("identity_conflict", "email or username already in use")
	case errors.Is(err, password.ErrBusy):
		return http.StatusServiceUnavailable, response.NewError("service_busy", "service temporarily busy")
	default:
		return http.StatusInternalServerError, response.NewError("internal_error", "internal server error")
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, data := mapError(err)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "handle request", "err", err)
	}

	if err := response.WriteJSON(w, data, status); err != nil {
		slog.ErrorContext(r.Context(), "write error response", "err", err)
	}
}
