package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"auth/internal/repository"
	"auth/internal/response"
)

// RegisterRequest is the JSON body for account registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,auth_email"`
	Username string `json:"username" validate:"required,min=3,max=30,auth_username"`
	Password string `json:"password" validate:"required,min=15,max=128"`
}

// RegisteredUser contains the safe user fields returned after registration.
type RegisteredUser struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
}

// RegisterResponse is the JSON body returned after registration.
type RegisterResponse struct {
	User RegisteredUser `json:"user"`
}

// Register creates a user account.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if err := decodeRequest(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Username = strings.TrimSpace(request.Username)

	if err := h.validate.Struct(&request); err != nil {
		writeError(w, r, ErrInvalidRequest)
		return
	}

	user, err := h.userService.Register(
		r.Context(),
		request.Email,
		request.Username,
		request.Password,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := response.WriteJSON(w, toRegisterResponse(user), http.StatusCreated); err != nil {
		slog.ErrorContext(r.Context(), "write registration response", "err", err)
	}
}

func toRegisterResponse(user repository.User) RegisterResponse {
	return RegisterResponse{
		User: RegisteredUser{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
	}
}
