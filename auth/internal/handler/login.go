package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"auth/internal/response"
	"auth/internal/service"
)

// LoginRequest is the JSON body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,auth_email"`
	Password string `json:"password" validate:"required"`
}

// LoginUser contains the safe user fields returned after login.
type LoginUser struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// LoginSession contains the session data returned after login.
type LoginSession struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// LoginResponse is the JSON body returned after login.
type LoginResponse struct {
	User    LoginUser    `json:"user"`
	Session LoginSession `json:"session"`
}

// Login authenticates a user and creates a session.
func (h *SessionHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if err := decodeRequest(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if err := h.validate.Struct(&request); err != nil {
		writeError(w, r, ErrInvalidRequest)
		return
	}

	result, err := h.sessionService.Login(
		r.Context(),
		request.Email,
		request.Password,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := response.WriteJSON(w, toLoginResponse(result), http.StatusOK); err != nil {
		slog.ErrorContext(r.Context(), "write login response", "err", err)
	}
}

func toLoginResponse(result service.LoginResult) LoginResponse {
	return LoginResponse{
		User: LoginUser{
			ID:       result.User.ID,
			Email:    result.User.Email,
			Username: result.User.Username,
		},
		Session: LoginSession{
			ID:        result.Session.ID,
			Token:     result.Session.Token,
			ExpiresAt: result.Session.ExpiresAt,
		},
	}
}
