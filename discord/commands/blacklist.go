package commands

import (
	"github.com/bwmarrin/discordgo"
	"time"
	"turtl/config"
	"turtl/db"
	"turtl/utils"
)

func blacklistCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd.")
		return
	}

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file to blacklist, big man.")
		return
	}

	alreadyBlacklisted, ok := db.CheckBlacklist(args[0])
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if alreadyBlacklisted {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That file is already blacklisted!")
		return
	}

	ok = db.AddToBlacklist(args[0])
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	existing, ok := db.CheckObjectsForBlacklistedFile(args[0])
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
