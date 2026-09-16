package main

import (
	"context"
	"fmt"
	"forum_backend/db"
	"forum_backend/helper"
	"forum_backend/server"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()

	seeding, err := helper.FlagHandling()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	//initialize database
	database, err := db.Init(seeding)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	srv, err := server.Server(ctx, database)
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errMsg := fmt.Sprintf("Server error: %v", err.Error())
		log.Fatal(errMsg)
	}
}
