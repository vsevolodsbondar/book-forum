package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"auth/internal/config"
	"auth/internal/db"
	"auth/internal/password"
	"auth/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "err", err)
		os.Exit(1)
	}
}

// run owns application startup so deferred cleanup runs before main exits on error.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	hasher, err := password.New(
		cfg.Argon2MemoryKiB,
		cfg.Argon2Iterations,
		cfg.Argon2Parallelism,
		cfg.Argon2MaxConcurrency,
	)
	if err != nil {
		return fmt.Errorf("create password hasher: %w", err)
	}

	database, err := db.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)

	return http.ListenAndServe(addr, server.New(database, hasher))
}
