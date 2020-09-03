package routes

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/joho/godotenv/autoload"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"turtl/db"
	"turtl/discord"
	"turtl/storage"
	"turtl/structs"
	"turtl/utils"
)

var maxMemory int64 = 50000000 // 50mb
var maxFilesPerUpload = 5

type finalResponse struct {
	Success bool                         `json:"success"`
	Files   []structs.FileUploadResponse `json:"files"`
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	contentLength := r.Header.Get("Content-Length")
	if contentLength == "" {
		w.WriteHeader(http.StatusLengthRequired)
		_, _ = w.Write([]byte(`No content was provided`))
		return
	}

	err := r.ParseMultipartForm(maxMemory)
	if utils.HandleError(err, "parsing form body") {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}
	defer func() {
		err := r.MultipartForm.RemoveAll()
		if utils.HandleError(err, "removing temp files from form") {
			return
		}
	}()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`No Authorization header provided`))
		return
	}

	var currentUser structs.User
	var userReqOk bool
	if strings.HasPrefix(authHeader, "Bearer ") {
		token, err := jwt.Parse(strings.TrimPrefix(authHeader, "Bearer "), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return []byte(os.Getenv("APP_SECRET_KEY")), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`Unexpected signing method`))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["apikey"] == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`API key not provided`))
				return
			}

			authHeader = claims["apikey"].(string)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`Forged token`))
			return
		}
	}

	currentUser, userReqOk = db.GetAccountFromAPIKey(authHeader)

	if !userReqOk {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}
	if currentUser.APIKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`Invalid API key`))
		return
	}

	member, err := discord.Client.GuildMember(os.Getenv("DISCORD_GUILD"), currentUser.DiscordID)
	if member == nil || member.User == nil || member.User.ID == "" || utils.HandleError(err, "checking if user is member") {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`Not in Discord server - ` + os.Getenv("DISCORD_INVITE")))
		return
	}

	domain := r.MultipartForm.Value["domain"][0]
	if !utils.BucketExists(storage.Buckets, domain) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`That domain isn't supported`))
		return
	}

	var rootDomain string
	var wildcard string

	if strings.Count(domain, ".") > 1 { // active wildcard
		domainSlice := strings.Split(domain, ".")
		wildcard = domainSlice[0]
		rootDomain = domainSlice[1] + "." + domainSlice[2]
	} else {
		rootDomain = domain
	}

	if len(wildcard) > 30 {
		w.WriteHeader(http.StatusRequestURITooLong)
		_, _ = w.Write([]byte(`Subdomain/wildcard too long (limit 30 letters)`))
		return
	}

	guild, err := discord.Client.Guild(os.Getenv("DISCORD_GUILD"))
	if utils.HandleError(err, "can't get guild") || guild == nil || guild.ID == "" || guild.Roles == nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal server error`))
		return
	}

	whitelistedID := utils.DoesRoleNameExist(rootDomain, guild.Roles)
	if whitelistedID != "" {
		if !utils.ArrayContains(member.Roles, whitelistedID) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`You are not allowed to use that domain`))
			return
		}
	}

	files, ok := r.MultipartForm.File["files[]"]
	if !ok || len(files) <= 0 {
		w.WriteHeader(http.StatusLengthRequired)
		_, _ = w.Write([]byte(`No files were provided`))
		return
	}
	if len(files) > maxFilesPerUpload {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(`Only 5 files can be uploaded at a time`))
		return
	}

	var responses []structs.FileUploadResponse
	for _, f := range files {
		if f.Size > currentUser.UploadLimit {
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusRequestEntityTooLarge,
				Name:    f.Filename,
				URL:     "",
				Info:    "File size limit is 100mb",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				_, _ = w.Write([]byte(`File size limit is 100mb`))
			}
			continue
		}

		cozybad := strings.Split(f.Filename, ".")
		extension := cozybad[len(cozybad)-1]

		generatedName, ok := db.GenerateNewFileName(extension, domain)
		if !ok {
			log.Println("Failed to generate file key within 5 attempts.")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "Failed to generate file key within 5 attempts.",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`Failed to generate file key within 5 attempts. Please try again or contact Polairr.`))
			}
			continue
		}

		contentType := "application/octet-stream"
		if ct := f.Header.Get("content-type"); ct != "" {
			contentType = ct
		}

		if len(generatedName) > 30 {
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusRequestEntityTooLarge,
				Name:    f.Filename,
				URL:     "",
				Info:    "File extension too long",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`File extension too long`))
			}
			continue
		}

		open, err := f.Open()
		if utils.HandleError(err, "opening file") {
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}

		md5H := md5.New()
		sha256H := sha256.New()
		tPath := filepath.Join(os.Getenv("APP_TEMP"), generatedName)
		tFile, err := os.Create(tPath)
		if utils.HandleError(err, "creating temp file") {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}

		fileWriter := io.MultiWriter(md5H, sha256H, tFile)
		_, err = io.Copy(fileWriter, open)
		_ = tFile.Close()
		if utils.HandleError(err, "copying file info") {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}

		md5Sum := md5H.Sum(nil)
		sha256Sum := sha256H.Sum(nil)

		md5String := hex.EncodeToString(md5Sum)
		sha256String := hex.EncodeToString(sha256Sum)

		blacklisted, ok := db.IsFileBlacklisted(sha256String)
		if !ok {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}
		if blacklisted && ok {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")

			_, _ = discord.Client.ChannelMessageSend(os.Getenv("DISCORD_ALERTS"), "<@492459066900348958>\n**Attempted Blacklisted Upload**\n\n**Uploader:** <@"+currentUser.DiscordID+">\n**MD5:** "+md5String+"\n**SHA256:** "+sha256String)

			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusNotAcceptable,
				Name:    f.Filename,
				URL:     "",
				Info:    "File is blacklisted",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`File is blacklisted`))
			}
			continue
		}

		existingFileURL, ok := db.DoesFileSumExist(md5String, sha256String, domain)
		if !ok {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}
		if existingFileURL != "" && ok {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusAlreadyReported,
				Name:    f.Filename,
				URL:     existingFileURL,
				Info:    "File already exists",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusAlreadyReported)
				_, _ = w.Write([]byte(`File already exists`))
			}
			continue
		}

		test, _ := f.Open()
		buf := make([]byte, f.Size)
		_, err = test.Read(buf)
		if utils.HandleError(err, "read to buffer") {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}

		_, err = storage.S3Service.PutObject(&s3.PutObjectInput{
			Body:               bytes.NewReader(buf),
			Bucket:             aws.String(rootDomain),
			Key:                aws.String(generatedName),
			ACL:                aws.String("public-read"),
			ContentType:        aws.String(contentType),
			ContentDisposition: aws.String("inline; filename=" + generatedName),
		})
		if utils.HandleError(err, "uploading") {
			err = os.Remove(tPath)
			_ = utils.HandleError(err, "removing file from path")
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "Failed to upload file",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`Failed to upload file`))
			}
			continue
		}

		err = os.Remove(tPath)
		_ = utils.HandleError(err, "removing file from path")

		_, err = db.DB.Exec("insert into objects values ($1, $2, $3, $4, $5, $6, $7, $8)", rootDomain, wildcard, generatedName, currentUser.DiscordID, time.Now().Unix(), strings.ToUpper(hex.EncodeToString(md5Sum)), strings.ToUpper(hex.EncodeToString(sha256Sum)), 0)
		if utils.HandleError(err, "insert object into psql") {
			responses = append(responses, structs.FileUploadResponse{
				Success: false,
				Status:  http.StatusInternalServerError,
				Name:    f.Filename,
				URL:     "",
				Info:    "500 - Internal Server Error",
			})
			if len(files) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`500 - Internal Server Error`))
			}
			continue
		}

		responses = append(responses, structs.FileUploadResponse{
			Success: true,
			Status:  http.StatusOK,
			Name:    f.Filename,
			URL:     "https://" + domain + "/" + generatedName,
			Info:    "",
		})
	}

	statusCode := http.StatusOK
	failed := 0

	for _, res := range responses {
		if !res.Success {
			failed += 1
			statusCode = http.StatusMultiStatus
		}
	}
	if failed == len(responses) {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	_ = json.NewEncoder(w).Encode(finalResponse{Success: true, Files: responses})
}
