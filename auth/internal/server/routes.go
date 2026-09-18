package server

import (
	"database/sql"
	"net/http"

	"auth/internal/handler"
	"auth/internal/password"
)

// New returns the router wrapped with request logging and panic recovery.
func New(database *sql.DB, hasher *password.Hasher) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /heartbeat", handler.Heartbeat)

	return Recovery(Logger(mux))
}
