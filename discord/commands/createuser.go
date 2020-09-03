package commands

import (
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strings"
	"turtl/db"
	"turtl/utils"
)

func createuserCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return
	}

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a user to add, idot")
		return
	}

	memberID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(args[0], "<@"), "!"), ">")
	member, err := s.GuildMember(os.Getenv("DISCORD_GUILD"), memberID)
	if member == nil || member.User == nil || member.User.ID == "" || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "They're not in this server")
		return
	}

	currentAccount, ok := db.GetAccountFromDiscord(member.User.ID)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if currentAccount.APIKey != "" && currentAccount.DiscordID != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "This person already has an account.")
		return
	}

	generated, ok := db.CreateUser(member)
	if generated == "" || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	dm, _ := s.UserChannelCreate(member.User.ID)
	_, err = s.ChannelMessageSend(dm.ID, "Your Turtl API key is: `"+generated+"`. Please do not lose it or give it to anyone else.\n\nYou can generate a .sxcu (configuration) file by going to <#737767470789820496> and typing `+sxcu` (your API kill will be filled automatically).")
	if err != nil {
		dm, err = s.UserChannelCreate(os.Getenv("OWNER_ID"))
		if utils.HandleError(err, "DMing yourself") {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, member.Mention()+"'s DMs are disabled. Their API key is: `"+generated+"`.")
	}

	err = s.GuildMemberRoleAdd(os.Getenv("DISCORD_GUILD"), member.User.ID, os.Getenv("DISCORD_REG_ROLE"))
	_ = utils.HandleError(err, "adding big boye role")

	_, _ = s.ChannelMessageSend(m.ChannelID, "User has been created")
}

func init() {
	RegisterCommand(&Command{
		Exec:     createuserCommand,
		Trigger:  "createuser",
		Aliases:  nil,
		Usage:    "createuser <user>",
		Desc:     "Grant an API key to someone",
		Disabled: false,
	})
}
