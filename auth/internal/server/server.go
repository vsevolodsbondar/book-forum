package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"auth/internal/handler"
	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/service"

	"github.com/go-playground/validator/v10"
)

// New returns the router wrapped with request logging and panic recovery.
func New(db *sql.DB, hasher *password.Hasher, validate *validator.Validate, sessionLifetime, sessionIdleTimeout time.Duration) (http.Handler, error) {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	userService := service.NewUserService(userRepo, hasher)
	sessionService, err := service.NewSessionService(userRepo, sessionRepo, hasher, sessionLifetime, sessionIdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("create session service: %w", err)
	}

	userHandler := handler.NewUserHandler(userService, validate)
	sessionHandler := handler.NewSessionHandler(sessionService, validate)

	mux := http.NewServeMux()
	registerRoutes(mux, userHandler, sessionHandler)

	return Recovery(Logger(mux)), nil
}
