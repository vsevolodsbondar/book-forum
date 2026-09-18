package server

import (
	"database/sql"
	"net/http"

	"auth/internal/handler"
	"auth/internal/password"
	"auth/internal/repository"
	"auth/internal/service"

	"github.com/go-playground/validator/v10"
)

// New returns the router wrapped with request logging and panic recovery.
func New(db *sql.DB, hasher *password.Hasher, validate *validator.Validate) http.Handler {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, hasher)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /heartbeat", handler.Heartbeat)
	mux.HandleFunc("POST /v1/register", handler.Register(userService, validate))

	return Recovery(Logger(mux))
}
