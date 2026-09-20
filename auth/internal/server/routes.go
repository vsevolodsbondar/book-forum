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
	// Repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Services
	userService := service.NewUserService(userRepo, hasher)
	sessionService, err := service.NewSessionService(userRepo, sessionRepo, hasher, sessionLifetime, sessionIdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("create session service: %w", err)
	}

	// Handlers
	userHandler := handler.NewUserHandler(userService, validate)
	sessionHandler := handler.NewSessionHandler(sessionService, validate)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /heartbeat", handler.Heartbeat)
	mux.HandleFunc("POST /v1/register", userHandler.Register)
	mux.HandleFunc("POST /v1/login", sessionHandler.Login)
	mux.HandleFunc("POST /v1/session/validate", sessionHandler.Validate)
	mux.HandleFunc("POST /v1/logout", sessionHandler.Logout)

	// Middleware
	return Recovery(Logger(mux)), nil
}
