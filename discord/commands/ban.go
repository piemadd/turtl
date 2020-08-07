package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/db"
	"turtl/utils"
)

func banCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return
	}

	args := UseArgs(m)
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a user and a reason, idot.")
		return
	}

	userID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	resArray := utils.RemoveIndex(0, args)
	reason := strings.Join(resArray, " ")

	err := s.GuildBanCreateWithReason(m.GuildID, userID, reason, 7)
	if utils.HandleError(err, "ban user") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
}

func init() {
	RegisterCommand(&Command{
		Exec:     banCommand,
		Trigger:  "ban",
		Aliases:  nil,
		Usage:    "ban <user> <reason>",
		Desc:     "Ban someone from the server",
		Disabled: false,
	})
}
