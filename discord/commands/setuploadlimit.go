package commands

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"turtl/config"
	"turtl/db"
)

func setuploadlimitCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return
	}

	args := UseArgs(m)
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a user and new size, lsoer.")
		return
	}

	memberID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	member, err := s.GuildMember(config.DISCORD_GUILD, memberID)
	if member == nil || member.User == nil || member.User.ID == "" || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "They're not in this server")
		return
	}

	megabytes, err := strconv.Atoi(args[1])
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That isn't a valid number.")
		return
	}

	account, ok := db.GetAccountFromDiscord(member.User)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if account.APIKey == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That user doesn't have an account.")
		return
	}

	bytes := 1000 * 1000 * megabytes
	ok = db.SetUserUploadLimit(account, bytes)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "User upload limit updated.")
}

func init() {
	RegisterCommand(&Command{
		Exec:     setuploadlimitCommand,
		Trigger:  "setuploadlimit",
		Aliases:  nil,
		Usage:    "setuploadlimit <@user> <mb>",
		Desc:     "Set a user's upload limit",
		Disabled: false,
	})
}
