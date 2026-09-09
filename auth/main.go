package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"auth/internal/config"
	"auth/internal/db"
	"auth/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	database, err := db.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer database.Close()

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)

	return http.ListenAndServe(addr, server.New(database))
}
