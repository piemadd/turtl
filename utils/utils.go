package utils

import (
	"bytes"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"strings"
	"turtl/config"
)

func HandleError(err error, loc string) bool {
	if err != nil {
		params := discordgo.WebhookParams{
			Content:   "<@492459066900348958>\n**ERROR**\n\nLocation: `" + loc + "`\nError:\n```" + err.Error() + "```",
			Username:  "turtl",
			AvatarURL: "http://i.turtl.cloud/turtl.png",
		}
		reqBody, err := json.Marshal(params)
		if err != nil {
			return true
		}

		_, err = http.Post(config.ALERTS_WEBHOOK, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			return true
		}
		return true
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
