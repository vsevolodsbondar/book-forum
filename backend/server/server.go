package server

import (
	"database/sql"
	"forum_backend/handler"
	"forum_backend/middleware"
	"forum_backend/repository"
	"forum_backend/service"
	"log"
	"net/http"
)

type Config struct {
	Host string
	Port string
}

func Server(db *sql.DB) (*http.Server, error) {
	mux := http.NewServeMux()

	cfg := Config{
		Host: "0.0.0.0",
		Port: "8080",
	}

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: middleware.Chain(mux),
	}

	dependencyWiring(mux, db)

	log.Println("Launching server at", srv.Addr)
	return srv, nil
}

func dependencyWiring(mux *http.ServeMux, db *sql.DB) *http.ServeMux {
	//comments wiring
	commentsRepo := repository.NewSQLiteCommentRepository(db)
	commentService := service.NewCommentService(commentsRepo)
	commentHandler := handler.NewCommentHandler(commentService)

	//users wiring
	usersRepo := repository.NewSQLiteUserRepository(db)
	userService := service.NewUserService(usersRepo)
	userHandler := handler.NewUserHandler(userService)

	RegisterRoutes(mux, commentHandler, userHandler)

	return mux
}
