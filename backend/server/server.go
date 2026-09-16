package server

import (
	"context"
	"database/sql"
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

	dependencyWiring(mux, db)

	log.Println("Launching server at", srv.Addr)
	return srv, nil
}

func dependencyWiring(mux *http.ServeMux, db *sql.DB) *http.ServeMux {
	commentsRepo := repository.NewSQLiteCommentRepository(db)
	commentService := service.NewCommentService(commentsRepo)
	commentHandler := handler.NewCommentHandler(commentService)

	RegisterRoutes(mux, commentHandler)

	return mux
}
