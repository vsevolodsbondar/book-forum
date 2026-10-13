package handler

import (
	"encoding/json"
	"fmt"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"net/http"
)

type AuthHandler struct {
	authClient client.AuthInterface
}

func NewAuthHandler(authClient client.AuthInterface) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}
func (h *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) error {
	sessionCookie, err := helper.ExtractSessionCookie(r)

	result, err := h.authClient.ValidateSession(r.Context(), sessionCookie)
	if err != nil {
		return fmt.Errorf("%w: session validation failed", custom_err.ErrSessionValidationFail)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.User.ID)
	return nil
}
