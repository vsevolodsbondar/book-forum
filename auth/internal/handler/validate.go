package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"auth/internal/response"
	"auth/internal/service"
)

// ValidationUser contains the user identity returned for a valid session.
type ValidationUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// ValidationSession contains the safe session fields returned after validation.
type ValidationSession struct {
	ID        string `json:"id"`
	ExpiresAt int64  `json:"expires_at"`
}

// ValidationResponse is the JSON body returned for a valid session.
type ValidationResponse struct {
	Session ValidationSession `json:"session"`
	User    ValidationUser    `json:"user"`
}

// Validate resolves an active bearer token to its session and user identity.
func (h *SessionHandler) Validate(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, r, service.ErrInvalidSession)
		return
	}

	result, err := h.sessionService.Validate(r.Context(), token)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := response.WriteJSON(w, toValidationResponse(result), http.StatusOK); err != nil {
		slog.ErrorContext(r.Context(), "write validation response", "err", err)
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

func toValidationResponse(result service.ValidationResult) ValidationResponse {
	return ValidationResponse{
		Session: ValidationSession{ID: result.SessionID, ExpiresAt: result.ExpiresAt},
		User:    ValidationUser{ID: result.UserID, Username: result.Username},
	}
}
