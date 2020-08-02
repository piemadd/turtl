package commands

import (
	"github.com/bwmarrin/discordgo"
	"turtl/db"
)

func reasonCommand(s *discordgo.Session, m *discordgo.Message) {
	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file to check, nerd.")
		return
	}

	blacklist, ok := db.GetBlacklist(args[0])
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if blacklist.SHA256 == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "File isn't blacklisted.")
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "**Blacklist Info**\n\n**SHA256:** "+blacklist.SHA256+"\n**Reason:** "+blacklist.Reason)
}

func init() {
	RegisterCommand(&Command{
		Exec:     reasonCommand,
		Trigger:  "reason",
		Aliases:  []string{"whyblacklist", "whyblacklisted"},
		Usage:    "reason <sha256 sum>",
		Desc:     "Get the reason for a file blacklist",
		Disabled: false,
	})
}
