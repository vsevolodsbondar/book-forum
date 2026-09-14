package server

import (
	"forum_backend/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, app *handler.Application) {
	mux.HandleFunc("GET /api/posts", app.GetLanding)
	mux.HandleFunc("GET /api/users", app.GetUser)
}
