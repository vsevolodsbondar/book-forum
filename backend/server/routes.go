package server

import (
	"forum_backend/custom_err"
	"forum_backend/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, comments *handler.CommentHandler) {
	mux.HandleFunc("GET /api/posts", handler.GetLanding)
	mux.HandleFunc("GET /posts/{id}/comments", custom_err.GlobalErrorHandler(comments.GetAllByPostID))
	mux.HandleFunc("POST /posts/{id}/comments", custom_err.GlobalErrorHandler(comments.Create))
	mux.HandleFunc("PATCH /comments/{id}", custom_err.GlobalErrorHandler(comments.Update))
	mux.HandleFunc("DELETE /comments/{id}", custom_err.GlobalErrorHandler(comments.Delete))
}
