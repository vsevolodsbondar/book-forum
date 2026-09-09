package server

import (
	"auth/internal/handler"
	"database/sql"
	"net/http"
)

// New returns the router wrapped with request logging and panic recovery.
func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /heartbeat", handler.Heartbeat)

	return Recovery(Logger(mux))
}
