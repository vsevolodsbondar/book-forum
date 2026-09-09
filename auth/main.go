package main

import (
	"log/slog"

	"auth/internal/config"
	"auth/internal/db"
)

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg.Database)
	if err != nil {
		slog.Error("opening database", "err", err)
		return
	}
	defer database.Close()

}
