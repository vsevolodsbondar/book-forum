package server

import (
	"forum_backend/custom_err"
	"forum_backend/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, comments *handler.CommentHandler, users *handler.UserHandler, posts *handler.PostHandler, auth *handler.AuthHandler) {
	// mux.HandleFunc("GET /api/posts", handler.GetLanding)
	mux.HandleFunc("GET /posts/{id}/comments", custom_err.GlobalErrorHandler(comments.GetAllByPostID))
	mux.HandleFunc("POST /posts/{id}/comments", custom_err.GlobalErrorHandler(comments.Create))
	mux.HandleFunc("PATCH /comments/{id}", custom_err.GlobalErrorHandler(comments.Update))
	mux.HandleFunc("DELETE /comments/{id}", custom_err.GlobalErrorHandler(comments.Delete))

	//users
	mux.HandleFunc("GET /users/{id}", custom_err.GlobalErrorHandler(users.GetUser))
	mux.HandleFunc("POST /users", custom_err.GlobalErrorHandler(users.CreateUser))
	mux.HandleFunc("PATCH /users/{id}", custom_err.GlobalErrorHandler(users.UpdateUser))
	mux.HandleFunc("DELETE /users/{id}", custom_err.GlobalErrorHandler(users.DeleteUser))

	//posts
	mux.HandleFunc("GET /posts", custom_err.GlobalErrorHandler(posts.GetPosts))
	//check for auth user
	mux.HandleFunc("GET /me", custom_err.GlobalErrorHandler(auth.AuthUser))
}
