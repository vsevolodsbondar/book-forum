package server

import (
	"forum_backend/custom_err"
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
	mux.HandleFunc("GET /users/{id}", custom_err.GlobalErrorHandler(users.GetUser))
	mux.HandleFunc("POST /users", custom_err.GlobalErrorHandler(users.CreateUser))
	mux.HandleFunc("PATCH /users/{id}", custom_err.GlobalErrorHandler(users.UpdateUser))
	mux.HandleFunc("DELETE /users/{id}", custom_err.GlobalErrorHandler(users.DeleteUser))
}
