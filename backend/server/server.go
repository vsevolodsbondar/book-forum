package server

import (
	"context"
	"forum_backend/handler"
	"forum_backend/middleware"
	"log"
	"net"
	"net/http"
)

type Config struct {
	Host string
	Port string
}

func Server(ctx context.Context, app *handler.Application) (*http.Server, error) {
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

	RegisterRoutes(mux, app)

	log.Println("Launching server at", srv.Addr)
	return srv, nil
}
