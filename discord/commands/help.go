package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/utils"
)

func helpCommand(s *discordgo.Session, m *discordgo.Message) {
	args := UseArgs(m)
	if len(args) > 0 {
		help := HelpManual(args[0])
		if help == "" {
			_, _ = s.ChannelMessageSend(m.ChannelID, "I couldn't find a command by that name!")
			return
		}

		_, _ = s.ChannelMessageSend(m.ChannelID, help)
		return
	}

	msg := "```md\n**Help Manual**\n\nRun `+help <command>` to get more information about a command.\n\n# Commands"

	for _, cmd := range CommandMap {
		msg += "\n+"+cmd.Trigger+": "+cmd.Desc
	}

	_, err := s.ChannelMessageSend(m.ChannelID, msg + "```")
	if utils.HandleError(err, "sending help message") {return}
}

func HelpManual(command string) string {
	cmd, ok := CommandMap[command]
	if !ok {
		cmd, ok = CommandMap[AliasMap[command]]
		if !ok {
			return ""
		}
	}

	usage := cmd.Usage

	msg := "```md\n**Help Manual: "+cmd.Trigger+"**\n\n# Description\n"+cmd.Desc+"\n\n# Usage\n"+usage+""

	if len(cmd.Aliases) > 0 {
		var aliases []string
		for _, alias := range cmd.Aliases {
			aliases = append(aliases, alias)
		}

		msg += "\n\n# Aliases\n"+strings.Join(aliases, ",")
	}

	return msg + "```"
}
func init() {
	RegisterCommand(&Command{
		Exec:       helpCommand,
		Trigger:    "help",
		Aliases:    []string{"?"},
		Usage:      "help [command]",
		Desc:       "Shows this message",
		Disabled:   false,
	})
}