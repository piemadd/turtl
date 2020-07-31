package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/config"
	"turtl/db"
	"turtl/utils"
)

func revokeCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return
	}

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a user to remove, idot")
		return
	}

	userID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	exists, ok := db.DoesDiscordOrKeyExist(userID)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if !exists {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That user doesn't exist")
		return
	}

	member, err := s.GuildMember(config.DISCORD_GUILD, userID)
	if member != nil && err == nil && member.User.ID != "" {
		err = s.GuildMemberRoleRemove(config.DISCORD_GUILD, member.User.ID, config.BIG_BOYE)
		_ = utils.HandleError(err, "removing big boye role")
	}

	ok = db.RevokeKey(userID)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, "User has been yoinked.")
	}
}

func init() {
	RegisterCommand(&Command{
		Exec:     revokeCommand,
		Trigger:  "revoke",
		Aliases:  []string{"yoink"},
		Usage:    "revoke <user/apikey>",
		Desc:     "Revoke an API key",
		Disabled: false,
	})
}
