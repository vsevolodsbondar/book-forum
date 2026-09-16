package server

import (
	"forum_backend/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, comments *handler.CommentHandler) {
	mux.HandleFunc("GET /api/posts", handler.GetLanding)
}
