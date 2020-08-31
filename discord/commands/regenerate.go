package commands

import (
	"github.com/bwmarrin/discordgo"
	"turtl/config"
	"turtl/db"
)

func regenerateCommand(s *discordgo.Session, m *discordgo.Message) {
	member, err := s.GuildMember(config.DISCORD_GUILD, m.Author.ID)
	if member == nil || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't find your account. Please DM Polairr to make one.")
		return
	}

	account, ok := db.GetAccountFromDiscord(member.User)
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
		ok = db.SetMemberAPIKey(member, generated)
		if !ok {
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
