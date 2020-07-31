package commands

import (
	"github.com/bwmarrin/discordgo"
)

func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func init() {
	RegisterCommand(&Command{
		Exec:       pingCommand,
		Trigger:    "ping",
		Aliases:    nil,
		Usage:      "ping",
		Desc:       "EEEEEEEEEEEE",
		Disabled:   false,
	})
}
