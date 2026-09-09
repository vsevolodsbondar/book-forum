package main

import (
	"context"
	"fmt"
	"forum_backend/server"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	srv, err := server.Server(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errMsg := fmt.Sprintf("Server error: %v", err.Error())
		log.Fatal(errMsg)
	}
}
