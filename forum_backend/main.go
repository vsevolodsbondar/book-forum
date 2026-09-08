package main

import (
	"fmt"
	"forum_backend/server"
	"log"
	"net/http"
)

func main() {

	fmt.Println("Starting server...")
	log.Fatalln(http.ListenAndServe(":8080", server.NewRouter()))
}
