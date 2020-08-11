package commands

import (
	"github.com/bwmarrin/discordgo"
	"turtl/db"
)

func deleteCommand(s *discordgo.Session, m *discordgo.Message) {
	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file to delete! Please use the file URL as the first argument.")
		return
	}

	file, ok := db.GetFileFromURL(args[0])
	if !ok || file.Bucket == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I couldn't find that file.")
		return
	}

	if file.DeletedAt != 0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "This file has already been deleted.")
		return
	}

	isAdmin, ok := db.CheckAdmin(m)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	if file.Uploader != m.Author.ID && !isAdmin {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That file doesn't belong to you!")
		return
	}

	ok = db.DeleteFile(file)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "File has been successfully deleted!")
}

func init() {
	RegisterCommand(&Command{
		Exec:     deleteCommand,
		Trigger:  "delete",
		Aliases:  []string{"deletefile", "remove", "removefile"},
		Usage:    "delete <file URL>",
		Desc:     "Delete a file from our servers",
		Disabled: false,
	})
}
