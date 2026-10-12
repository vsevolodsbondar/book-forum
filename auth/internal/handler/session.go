package handler

import (
	"auth/internal/service"

	"github.com/go-playground/validator/v10"
)

// SessionHandler handles login and session lifecycle requests.
type SessionHandler struct {
	sessionService *service.SessionService
	validate       *validator.Validate
}

// NewSessionHandler creates a session handler.
func NewSessionHandler(sessionService *service.SessionService, validate *validator.Validate) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		validate:       validate,
	}
}
