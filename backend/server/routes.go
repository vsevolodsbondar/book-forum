package server

import (
	"forum_backend/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, comments *handler.CommentHandler, users *handler.UserHandler) {
	// mux.HandleFunc("GET /api/posts", handler.GetLanding)
	mux.HandleFunc("GET /posts/{id}/comments", comments.GetAllByPostID)
	mux.HandleFunc("POST /posts/{id}/comments", comments.Create)
	mux.HandleFunc("PATCH /comments/{id}", comments.Update)
	mux.HandleFunc("DELETE /comments/{id}", comments.Delete)
	//users
	mux.HandleFunc("GET /users/{id}", users.GetUser)
	mux.HandleFunc("POST /users", users.CreateUser)
	mux.HandleFunc("PATCH /users/{id}", users.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", users.DeleteUser)
}
