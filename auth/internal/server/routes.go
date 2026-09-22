package server

import (
	"net/http"

	"auth/internal/handler"
)

func registerRoutes(mux *http.ServeMux, users *handler.UserHandler, sessions *handler.SessionHandler) {
	mux.HandleFunc("GET /heartbeat", handler.Heartbeat)

	mux.HandleFunc("POST /v1/register", users.Register)
	mux.HandleFunc("POST /v1/login", sessions.Login)
	mux.HandleFunc("POST /v1/session/validate", sessions.Validate)
	mux.HandleFunc("POST /v1/logout", sessions.Logout)
}
