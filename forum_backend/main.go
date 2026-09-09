package main

import (
	"fmt"
	"forum_backend/db"
	"forum_backend/server"
	"log"
	"net/http"
)

func main() {
	//initialize database
	database, err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	fmt.Println("Starting server...")
	log.Fatalln(http.ListenAndServe(":8080", server.NewRouter()))
}
