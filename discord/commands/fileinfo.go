package commands

import (
	"github.com/bwmarrin/discordgo"
	"time"
	"turtl/db"
)

func fileinfoCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd.")
		return
	}

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file, big man.")
		return
	}

	file, ok := db.GetFileFromURL(args[0])
	if !ok || file.Bucket == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I couldn't find that file.")
		return
	}

	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil || dm == nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Enable your DMs, nerd.")
		return
	}

	_, err = s.ChannelMessageSend(dm.ID, "**Upload Info: "+args[0]+"**\n\n**Uploader:** <@"+file.Uploader+">\n**Created At:** "+time.Unix(int64(file.CreatedAt), 0).Format(time.RFC1123)+"\n**MD5:** "+file.MD5+"\n**SHA256:** "+file.SHA256)
}

func init() {
	RegisterCommand(&Command{
		Exec:     fileinfoCommand,
		Trigger:  "fileinfo",
		Aliases:  []string{"info", "uploadinfo"},
		Usage:    "fileinfo <file URL>",
		Desc:     "Get info about a file",
		Disabled: false,
	})
}
