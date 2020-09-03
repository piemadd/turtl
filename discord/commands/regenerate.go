package commands

import (
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"turtl/db"
	"turtl/utils"
)

func regenerateCommand(s *discordgo.Session, m *discordgo.Message) {
	member, err := s.GuildMember(os.Getenv("DISCORD_GUILD"), m.Author.ID)
	if member == nil || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't find your account. Please DM Polairr to make one.")
		return
	}

	account, ok := db.GetAccountFromDiscord(member.User.ID)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if account.DiscordID == "" || account.APIKey == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't find your account. Please DM Polairr to make one.")
		return
	}

	generated, ok := db.GenerateUUID()
	if !ok || generated == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	dm, err := s.UserChannelCreate(member.User.ID)
	if dm == nil || err != nil || dm.ID == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Please make sure your DMs are open and try running this command again.")
		return
	}
	msg, err := s.ChannelMessageSend(dm.ID, "Your Turtl API key is: `"+generated+"`. Please do not lose it or give it to anyone else.\n\nYou can generate a .sxcu (configuration) file by going to <#737767470789820496> and typing `+sxcu` (your API kill will be filled automatically).")
	if err != nil || msg == nil || msg.ID == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Please make sure your DMs are open and try running this command again.")
		return
	} else {
		_, err := db.DB.Exec("update users set apikey=$1 where discordid=$2", generated, member.User.ID)
		if utils.HandleError(err, "update api key") {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
			_, _ = s.ChannelMessageEdit(msg.ChannelID, msg.ID, "Error! Please try again later.")
			return
		}

		_, _ = s.ChannelMessageSend(m.ChannelID, "Check your DMs!")
	}
}

func init() {
	RegisterCommand(&Command{
		Exec:     regenerateCommand,
		Trigger:  "regenerate",
		Aliases:  nil,
		Usage:    "regenerate",
		Desc:     "Regenerate your API key",
		Disabled: false,
	})
}
