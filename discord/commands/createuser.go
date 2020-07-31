package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/config"
	"turtl/db"
	"turtl/utils"
)

func createuserCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(s, m)
	if !allowed || !ok { return }

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a user to add, idot")
		return
	}

	memberID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	member, err := s.GuildMember(config.DISCORD_GUILD, memberID)
	if member == nil || member.User == nil || member.User.ID == "" || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "They're not in this server")
		return
	}

	generated, ok := db.CreateUser(s, m, member)

	dm, err := s.UserChannelCreate(member.User.ID)
	if utils.HandleError(err, "opening dm") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	_, err = s.ChannelMessageSend(dm.ID, "Your Turtl API key is: `"+generated+"`. Please do not lose it or give it to anyone else.\n\nYou can generate a .sxcu (configuration) file by going to <#737767470789820496> and typing `+sxcu` (your API kill will be filled automatically).")
	if err != nil {
		dm, err = s.UserChannelCreate(config.POLAIRR_ID)
		if utils.HandleError(err, "DMing yourself") {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, member.Mention()+"'s DMs are disabled. Their API key is: `"+generated+"`.")
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "User has been created")
}

func init() {
	RegisterCommand(&Command{
		Exec:       createuserCommand,
		Trigger:    "createuser",
		Aliases:    nil,
		Usage:      "createuser <user>",
		Desc:       "Grant an API key to someone",
		Disabled:   false,
	})
}
