package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/parnurzeal/gorequest"
	"os"
	"strings"
)

var AppSecretKey []byte

func HandleError(err error, loc string) bool {
	if err != nil {
		params := discordgo.WebhookParams{
			Content:   "<@492459066900348958>\n**ERROR**\n\nLocation: `" + loc + "`\nError:\n```" + err.Error() + "```",
			Username:  "turtl",
			AvatarURL: "https://i.turtl.cloud/turtl.png",
		}

		_, _, _ = gorequest.New().Post(os.Getenv("DISCORD_ALERT_WEBHOOK")).
			Send(params).
			End()
		return true
	}
	return false
}

func RemoveIndex(index int, arr []string) []string {
	copy(arr[index:], arr[index+1:])
	arr[len(arr)-1] = ""
	arr = arr[:len(arr)-1]

	return arr
}

func BucketExists(arr []*s3.Bucket, query string) bool {
	var queryString string
	if strings.Count(query, ".") > 1 {
		eeee := strings.Split(query, ".")
		queryString = strings.Join(RemoveIndex(0, eeee), ".")
	} else {
		queryString = query
	}
	for _, e := range arr {
		if strings.ToLower(aws.StringValue(e.Name)) == strings.ToLower(queryString) {
			return true
		} else {
			continue
		}
	}

	return false
}

func ArrayContains(arr []string, query string) bool {
	for _, e := range arr {
		if strings.ToLower(e) == strings.ToLower(query) {
			return true
		} else {
			continue
		}
	}

	return false
}

func DoesRoleNameExist(name string, roles []*discordgo.Role) string {
	for _, r := range roles {
		if strings.ToLower(name) == strings.ToLower(r.Name) {
			return r.ID
		}
	}

	return ""
}
