package db

import (
	"database/sql"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var DB *sql.DB

func init() {
	var err error
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_DB"), os.Getenv("PG_SSL"))

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("failed to connect to postgres ", err.Error())
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("failed to ping postgres ", err.Error())
	}

	log.Println("Successfully connected to PostgreSQL database.")
}
