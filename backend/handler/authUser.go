package handler

import (
	"encoding/json"
	"fmt"
	"forum_backend/client"
	"forum_backend/custom_err"
	"net/http"
)

type AuthHandler struct {
	authClient client.AuthInterface
}

func NewAuthHandler(authClient client.AuthInterface) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}
func (h *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) error {
	session, err := r.Cookie("session")
	if err != nil {
		return fmt.Errorf("%w: no session cookie", custom_err.ErrExpiredSession)
	}
	result, err := h.authClient.ValidateSession(r.Context(), session.Value)
	if err != nil {
		return fmt.Errorf("%w: session validation failed", custom_err.ErrSessionValidationFail)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.User.ID)
	return nil
}
