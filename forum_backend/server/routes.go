package server

import (
	"forum_backend/handler"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handler.GetLanding)
	//middlewares

	return mux
}
