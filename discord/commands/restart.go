package commands

import (
	"github.com/bwmarrin/discordgo"
	"os"
	"turtl/db"
)

func restartCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd.")
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "cya, nerds")
	os.Exit(1)
}

func init() {
	RegisterCommand(&Command{
		Exec:     restartCommand,
		Trigger:  "restart",
		Aliases:  nil,
		Usage:    "restart",
		Desc:     "Restart the entire service (kills the process and gets restarted by Docker)",
		Disabled: false,
	})
}
