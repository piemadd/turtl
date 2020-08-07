package db

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"turtl/storage"
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
		err = objects.Scan(&existingObject.Bucket, &existingObject.Wildcard, &existingObject.FileName, &existingObject.Uploader, &existingObject.CreatedAt, &existingObject.MD5, &existingObject.SHA256, &existingObject.DeletedAt)
		if existingObject.Wildcard == "" {
			return "https://" + existingObject.Bucket + "/" + existingObject.FileName, true
		} else {
			return "https://" + existingObject.Wildcard + "." + existingObject.Bucket + "/" + existingObject.FileName, true
		}
	}
	return "", true
}

func GetFileFromURL(url string) (structs.Object, bool) {
	if url == "" {
		return structs.Object{}, false
	}
	if strings.Count(url, ".") < 2 {
		return structs.Object{}, false
	}

	url = strings.TrimPrefix(url, "https://")
	splitAtPeriods := strings.Split(url, ".")
	splitAtSlash := strings.Split(url, "/")
	filename := splitAtSlash[1]

	var wildcard string
	var domain string
	if len(splitAtPeriods) == 2 { // no wildcard
		domain = splitAtPeriods[0] + "." + strings.TrimSuffix(splitAtPeriods[1], "/"+strings.Split(filename, ".")[0])
		wildcard = ""
	} else {
		domain = splitAtPeriods[1] + "." + strings.TrimSuffix(splitAtPeriods[2], "/"+strings.Split(filename, ".")[0])
		wildcard = splitAtPeriods[0]
	}

	fmt.Println(filename)
	fmt.Println(domain)
	fmt.Println(wildcard)

	rows, err := DB.Query("select * from objects where wildcard=$1 and bucket=$2 and filename=$3", wildcard, domain, filename)
	if utils.HandleError(err, "query DB for GetFileFromURL") {
		return structs.Object{}, false
	}
	defer rows.Close()
	if rows.Next() {
		var retVal structs.Object
		err = rows.Scan(&retVal.Bucket, &retVal.Wildcard, &retVal.FileName, &retVal.Uploader, &retVal.CreatedAt, &retVal.MD5, &retVal.SHA256, &retVal.DeletedAt)
		if utils.HandleError(err, "scan into retval at GetFileFromURL") {
			return structs.Object{}, false
		}
		return retVal, true
	}
	return structs.Object{}, true
}

func GetFileFromHash(sha256 string) (structs.Object, bool) {
	rows, err := DB.Query("select * from objects where sha256=$1", sha256)
	if utils.HandleError(err, "query DB for GetFileFromHash") {
		return structs.Object{}, false
	}
	defer rows.Close()
	if rows.Next() {
		var retVal structs.Object
		err = rows.Scan(&retVal.Bucket, &retVal.Wildcard, &retVal.FileName, &retVal.Uploader, &retVal.CreatedAt, &retVal.MD5, &retVal.SHA256, &retVal.DeletedAt)
		if utils.HandleError(err, "scan into retval at GetFileFromHash") {
			return structs.Object{}, false
		}
		return retVal, true
	}
	return structs.Object{}, true
}

func CheckObjectsForBlacklistedFile(sha256 string) ([]structs.Object, bool) {
	rows, err := DB.Query("select * from objects where sha256=$1", sha256)
	if utils.HandleError(err, "check objects for blacklisted files") {
		return []structs.Object{}, false
	}
	defer rows.Close()
	var retVal []structs.Object
	for rows.Next() {
		var t structs.Object
		err = rows.Scan(&t.Bucket, &t.Wildcard, &t.FileName, &t.Uploader, &t.CreatedAt, &t.MD5, &t.SHA256, &t.DeletedAt)
		if utils.HandleError(err, "scan into retval at CheckObjectsForBlacklistedFile") {
			return []structs.Object{}, false
		}
		if t.DeletedAt == 0 {
			continue
		}

		retVal = append(retVal, t)
	}
	return retVal, true
}

func GetBlacklist(sha256 string) (structs.Blacklist, bool) {
	rows, err := DB.Query("select * from blacklist where hash=$1", sha256)
	if utils.HandleError(err, "check blacklist") {
		return structs.Blacklist{}, false
	}
	defer rows.Close()
	if rows.Next() {
		var retVal structs.Blacklist
		err = rows.Scan(&retVal.SHA256, &retVal.Reason)
		if utils.HandleError(err, "scan blacklist info into retval") {
			return structs.Blacklist{}, false
		}
		return retVal, true
	}
	return structs.Blacklist{}, true
}

func AddToBlacklist(sha256 string, reason string) bool {
	_, err := DB.Exec("insert into blacklist values ($1, $2)", sha256, reason)
	if utils.HandleError(err, "inserting hash into blacklist") {
		return false
	}
	return true
}

func DeleteFile(file structs.Object) bool {
	_, err := storage.S3Service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(file.Bucket),
		Key:    aws.String(file.FileName),
	})
	if utils.HandleError(err, "delete file from bucket") {
		return false
	}

	_, err = DB.Exec("update objects set deletedat=$1 where bucket=$2 and wildcard=$3 and filename=$4 and uploader=$5 and createdat=$6 and md5=$7 and sha256=$8", time.Now().Unix(), file.Bucket, file.Wildcard, file.FileName, file.Uploader, file.CreatedAt, file.MD5, file.SHA256)
	if utils.HandleError(err, "delete file from db") {
		return false
	}

	return true
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

func CheckAdmin(m *discordgo.Message) (bool, bool) {
	users, err := DB.Query("select * from users where discordid=$1 and admin=true", m.Author.ID)
	if utils.HandleError(err, "query users to check admin") {
		return false, false
	}
	defer users.Close()
	if !users.Next() {
		return false, true
	}
	return true, true
}

func CreateUser(s *discordgo.Session, m *discordgo.Message, member *discordgo.Member) (string, bool) {
	generated, ok := GenerateUUID(s, m)
	if !ok || generated == "" {
		return "", false
	}

	_, err := DB.Exec("insert into users values ($1, $2, 100000000, false)", member.User.ID, generated)
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

func RevokeKey(key string) bool {
	_, err := DB.Exec("delete from users where discordid=$1 or apikey=$1", key)
	if utils.HandleError(err, "delete user from db") {
		return false
	}
	return true
}

func DoesDiscordOrKeyExist(key string) (bool, bool) {
	rows, err := DB.Query("select * from users where discordid=$1 or apikey=$1", key)
	if utils.HandleError(err, "delete user from db") {
		return false, false
	}
	defer rows.Close()
	if rows.Next() {
		return true, true
	}
	return false, true
}

func SetMemberAPIKey(member *discordgo.Member, newAPIKey string) bool {
	_, err := DB.Exec("update users set apikey=$1 where discordid=$2", newAPIKey, member.User.ID)
	if utils.HandleError(err, "update api key") {
		return false
	}
	return true
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
		err = users.Scan(&retUser.DiscordID, &retUser.APIKey, &retUser.UploadLimit, &retUser.Admin)
		if utils.HandleError(err, "get discord member account scan") {
			return structs.User{}, false
		}
		return retUser, true
	}
	return structs.User{}, true
}

func SetUserUploadLimit(user structs.User, newLimit int) bool {
	_, err := DB.Exec("update users set uploadlimit=$1 where discordid=$2 and apikey=$3", newLimit, user.DiscordID, user.APIKey)
	if utils.HandleError(err, "updating upload limit in db") {
		return false
	}
	return true
}
