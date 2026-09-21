package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/handler"
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

	validate, err := handler.NewValidator()
	if err != nil {
		return fmt.Errorf("create request validator: %w", err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	router, err := server.New(db, hasher, validate, cfg.SessionLifetime, cfg.SessionIdleTimeout)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)

	return http.ListenAndServe(addr, router)
}
