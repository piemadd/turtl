package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/config"
	"turtl/db"
	"turtl/structs"
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

	identifier := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	rows, err := db.DB.Query("select * from users where discordid=$1 or apikey=$1", identifier)
	if utils.HandleError(err, "delete user from db") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	defer rows.Close()
	if !rows.Next() {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That user doesn't exist")
		return
	}

	var account structs.User
	err = rows.Scan(&account.DiscordID, &account.APIKey, &account.UploadLimit, &account.Admin)
	if utils.HandleError(err, "scan account info in revoke.go") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	member, err := s.GuildMember(config.DISCORD_GUILD, account.DiscordID)
	if member != nil && err == nil && member.User.ID != "" {
		err = s.GuildMemberRoleRemove(config.DISCORD_GUILD, member.User.ID, config.BIG_BOYE)
		_ = utils.HandleError(err, "removing big boye role")
	}

	ok = db.RevokeKey(account.APIKey)
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
		Aliases:  []string{"yoink", "yeet"},
		Usage:    "revoke <user/apikey>",
		Desc:     "Revoke an API key",
		Disabled: false,
	})
}
