package main

import (
	"fmt"
	"forum_backend/db"
	"forum_backend/server"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	seeding := helper.IsSeeding()

	//initialize database
	database, err := db.Init()
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
