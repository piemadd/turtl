package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	_ "turtl/db"
	"turtl/discord"
	"turtl/routes"
	_ "turtl/storage"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/upload", routes.UploadFile)

	log.Println("API Initialized")

	go discord.CreateBot()

	err := http.ListenAndServe(":80", router)
	if err != nil {
		log.Fatal("Failed to start API", err.Error())
	}
}
