package handler

import (
	"auth/internal/service"

	"github.com/go-playground/validator/v10"
)

// UserHandler handles user account requests.
type UserHandler struct {
	userService *service.UserService
	validate    *validator.Validate
}

// NewUserHandler creates a user handler.
func NewUserHandler(userService *service.UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validate,
	}
}
