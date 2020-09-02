package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
	"turtl/config"
	"turtl/db"
	"turtl/structs"
	"turtl/utils"
)

func blacklistCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd.")
		return
	}

	args := UseArgs(m)
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file to blacklist and a reason, big man.")
		return
	}

	shaSum := args[0]
	reason := strings.Join(utils.RemoveIndex(0, args), " ")

	alreadyBlacklisted, ok := db.IsFileBlacklisted(shaSum)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if alreadyBlacklisted {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That file is already blacklisted!")
		return
	}

	_, err := db.DB.Exec("insert into blacklist values ($1, $2)", strings.ToUpper(shaSum), reason)
	if utils.HandleError(err, "inserting hash into blacklist") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	rows, err := db.DB.Query("select * from objects where sha256=$1", strings.ToUpper(shaSum))
	if utils.HandleError(err, "check objects for blacklisted files") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	defer rows.Close()
	var existing []structs.Object
	for rows.Next() {
		var t structs.Object
		err = rows.Scan(&t.Bucket, &t.Wildcard, &t.FileName, &t.Uploader, &t.CreatedAt, &t.MD5, &t.SHA256, &t.DeletedAt)
		if utils.HandleError(err, "scan into retval at CheckObjectsForBlacklistedFile") {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
			return
		}
		if t.DeletedAt == 0 {
			continue
		}

		existing = append(existing, t)
	}
	for _, file := range existing {
		_, err := s.ChannelMessageSend(config.DISCORD_ALERTS, "<@492459066900348958>\n**Blacklisted file in Storage**\n\n**Bucket:** "+file.Bucket+"\n**Wildcard:** "+file.Wildcard+"\n**File Name:** "+file.FileName+"\n**Uploader:** <@"+file.Uploader+">\n**Created At:**"+time.Unix(int64(file.CreatedAt), 0).Format(time.RFC1123)+"\n**MD5:** "+file.MD5+"\n**SHA256:** "+file.SHA256)
		if utils.HandleError(err, "send alert for blacklisted file") {
			continue
		}
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "File has been blacklisted")
}

func init() {
	RegisterCommand(&Command{
		Exec:     blacklistCommand,
		Trigger:  "blacklist",
		Aliases:  nil,
		Usage:    "blacklist <sha256 hash>",
		Desc:     "",
		Disabled: false,
	})
}
