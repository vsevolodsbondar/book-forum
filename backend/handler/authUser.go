package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/service"
	"net/http"
)

type AuthHandler struct {
	authClient  client.AuthInterface
	userService *service.UserService
}

func NewAuthHandler(authClient client.AuthInterface, userService *service.UserService) *AuthHandler {
	return &AuthHandler{authClient: authClient, userService: userService}
}

func (h *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}

	result, err := h.authClient.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return fmt.Errorf("%w: session validation failed", custom_err.ErrSessionValidationFail)
	}

	// the session is valid in auth, but the user must exist in our database too
	if _, err := h.userService.GetUser(ctx, result.User.ID); err != nil {
		if errors.Is(err, custom_err.ErrUserNotFound) {
			// a "broken" session: delete it in auth and clear the cookie
			_ = h.authClient.LogoutUser(ctx, sessionCookie)
			helper.ClearSessionCookie(w)
			return fmt.Errorf("%w: user does not exist", custom_err.ErrSessionValidationFail)
		}
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(result.User.ID)
}
