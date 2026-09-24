package server

import (
	"context"
	"database/sql"
	"forum_backend/client"
	"forum_backend/handler"
	"forum_backend/middleware"
	"forum_backend/repository"
	"forum_backend/service"
	"log"
	"net"
	"net/http"
)

type Config struct {
	Host string
	Port string
}

func Server(ctx context.Context, db *sql.DB) (*http.Server, error) {
	mux := http.NewServeMux()
	cfg := Config{
		Host: "0.0.0.0",
		Port: "8080",
	}
	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: middleware.Chain(mux),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	// later will inject true client
	// authClient := &client.AuthHTTPClient{
	// 	BaseURL: "http://auth:8081",
	// 	Client: &http.Client{
	// 		Timeout: 3 * time.Second,
	// 	},
	// }

	mockAuthClient := &client.MockAuthClient{}

	dependencyWiring(mux, db, mockAuthClient)

	log.Println("Launching server at", srv.Addr)
	return srv, nil
}

func dependencyWiring(mux *http.ServeMux, db *sql.DB, auth client.AuthInterface) *http.ServeMux {
	mux.HandleFunc("GET /api/posts", handler.GetLanding)
	commentsRepo := repository.NewSQLiteCommentRepository(db)
	commentService := service.NewCommentService(commentsRepo)
	commentHandler := handler.NewCommentHandler(commentService, auth)
	//users wiring
	usersRepo := repository.NewSQLiteUserRepository(db)
	userService := service.NewUserService(usersRepo, auth)
	userHandler := handler.NewUserHandler(userService, auth)

	authSession := handler.NewAuthHandler(auth)
	RegisterRoutes(mux, commentHandler, userHandler, authSession)

	return mux
}
