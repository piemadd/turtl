package main

import (
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
	"os"
	_ "turtl/db"
	"turtl/discord"
	"turtl/routes"
	_ "turtl/storage"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/upload", routes.UploadFile)
	router.HandleFunc("/auth", routes.DiscordAuth)

	log.Println("API Initialized")

	go discord.CreateBot()

	err := http.ListenAndServeTLS(":443", os.Getenv("SSL_CERT"), os.Getenv("SSL_PRIV"), router)
	if err != nil {
		log.Fatal("Failed to start API", err.Error())
	}
}
