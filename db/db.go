package db

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"os"
	"time"
	"turtl/structs"
	"turtl/utils"
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

func IsFileBlacklisted(sha256 string) (bool, bool) {
	blacklist, err := DB.Query("select * from blacklist where hash=$1", sha256)
	if utils.HandleError(err, "check if file is blacklisted") {
		return true, false
	}
	defer blacklist.Close()
	if blacklist.Next() {
		return true, true
	}
	return false, true
}

func DoesFileSumExist(md5 string, sha256 string, domain string) (string, bool) {
	objects, err := DB.Query("select * from objects where (md5=$1 or sha256=$2) and bucket=$3", md5, sha256, domain)
	if utils.HandleError(err, "check if file sum exists") {
		return "", false
	}
	defer objects.Close()
	if objects.Next() {
		var existingObject structs.Object
		err = objects.Scan(&existingObject.Bucket, &existingObject.Wildcard, &existingObject.FileName, &existingObject.Uploader, &existingObject.CreatedAt, &existingObject.MD5, &existingObject.SHA256)
		return "http://" + existingObject.Wildcard + "." + existingObject.Bucket + "/" + existingObject.FileName, true
	}
	return "", true
}

func DoesFileNameExist(name string, domain string) (bool, bool) {
	rows, err := DB.Query("select * from objects where filename=$1 and bucket=$2", name, domain)
	if utils.HandleError(err, "query psql for file") {
		return true, false
	}
	defer rows.Close()
	if rows.Next() {
		return true, true
	}
	return false, true
}

func GenerateNewFileName(extension string, domain string) (string, bool) {
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for a := 0; a < 5; a++ {
		rand.Seed(time.Now().UnixNano())
		b := make([]byte, 10)
		for i := range b {
			b[i] = characters[rand.Intn(len(characters))]
		}

		formatted := string(b) + "." + extension
		exists, ok := DoesFileNameExist(formatted, domain)
		if !ok {
			return "", false
		}
		if !exists {
			return formatted, true
		}
	}
	return "", false
}

func CheckAdmin(s *discordgo.Session, m *discordgo.Message) (bool, bool) {
	users, err := DB.Query("select * from users where discordid=$1 and admin=true", m.Author.ID)
	if utils.HandleError(err, "query users to check admin") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return false, false
	}
	defer users.Close()
	if users.Next() {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return false, true
	}
	return true, true
}

func CreateUser(s *discordgo.Session, m *discordgo.Message, member *discordgo.Member) (string, bool) {
	generated, ok := GenerateUUID(s, m)
	if !ok || generated == "" {
		return "", false
	}

	_, err := DB.Exec("insert into users values ($1, $2, false)", member.User.ID, generated)
	if utils.HandleError(err, "query users to check for existing uuid") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return "", false
	}

	return generated, true
}

func GenerateUUID(s *discordgo.Session, m *discordgo.Message) (string, bool) {
	var generated uuid.UUID
	for i := 0; i < 5; i++ {
		generated = uuid.New()
		exists, ok := DoesUserExist(s, m, generated.String())
		if !ok {
			return "", false
		}
		if !exists {
			return generated.String(), true
		}
	}
	return "", false
}

func DoesUserExist(s *discordgo.Session, m *discordgo.Message, apikey string) (bool, bool) {
	existing, err := DB.Query("select * from users where apikey=$1", apikey)
	if utils.HandleError(err, "query users to check for existing uuid") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return false, false
	}
	defer existing.Close()
	if existing.Next() {
		return true, true
	}

	return false, true
}

func GetDiscordMemberAccount(member *discordgo.Member) (structs.User, bool) {
	users, err := DB.Query("select * from users where discordid=$1", member.User.ID)
	if utils.HandleError(err, "checking for discord account") {
		return structs.User{}, false
	}
	defer users.Close()
	if users.Next() {
		var retUser structs.User
		err = users.Scan(&retUser.DiscordID, &retUser.APIKey, &retUser.Admin)
		if utils.HandleError(err, "get discord member account scan") {
			return structs.User{}, false
		}
		return retUser, true
	}
	return structs.User{}, true
}
