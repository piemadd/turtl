package main

import (
	"crypto/tls"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	_ "turtl/db"
	"turtl/discord"
	"turtl/routes"
	_ "turtl/storage"
	"turtl/utils"
)

func main() {
	// generate secret key before we do anything else
	rand.Seed(time.Now().UnixNano())
	n := 4096
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*_-=+[]{}(),.<>/?;:~"

	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(letterBytes[rand.Intn(len(letterBytes))])
	}

	utils.AppSecretKey = []byte(sb.String())

	// start the bot
	go discord.CreateBot()

	// initialize API
	router := mux.NewRouter()

	router.HandleFunc("/upload", routes.UploadFile)
	router.HandleFunc("/auth", routes.DiscordAuth)

	tlsCfg := &tls.Config{
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	server := &http.Server{
		Addr:         ":443",
		Handler:      router,
		TLSConfig:    tlsCfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	log.Println("API Initialized. Starting...")

	err := server.ListenAndServeTLS(os.Getenv("SSL_CERT"), os.Getenv("SSL_PRIV"))
	if err != nil {
		log.Fatal("Failed to start API ", err.Error())
	}
}
