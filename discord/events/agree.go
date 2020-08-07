package events

import (
	"github.com/bwmarrin/discordgo"
	"turtl/config"
	"turtl/db"
	"turtl/utils"
)

func Agree(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	if e.ChannelID != "737854426693369899" {
		return
	}
	if e.MessageID != "741298295405543427" {
		err := s.MessageReactionRemove(e.ChannelID, e.MessageID, e.Emoji.ID, e.UserID)
		if utils.HandleError(err, "removing emoji from wrong message") {
			_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
			return
		}
	}
	if e.Emoji.ID != ":checkmark:741299438903099421" {
		err := s.MessageReactionRemove(e.ChannelID, e.MessageID, e.Emoji.ID, e.UserID)
		if utils.HandleError(err, "removing wrong emoji from agree message") {
			_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
			return
		}
	}

	err := s.MessageReactionRemove(e.ChannelID, e.MessageID, e.Emoji.ID, e.UserID)
	if utils.HandleError(err, "removing emoji from agree message") {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}

	dm, _ := s.UserChannelCreate(e.UserID)
	ee, err := s.ChannelMessageSend(dm.ID, "Creating account...")
	if ee == nil || ee.ID == "" || err != nil {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> Please enable your DMs, then try reacting to the message again.")
		return
	}

	member, err := s.GuildMember(config.DISCORD_GUILD, e.UserID)
	if member == nil || member.User == nil || member.User.ID == "" || utils.HandleError(err, "checking for member in agree") {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}

	acc, ok := db.GetDiscordMemberAccount(member)
	if !ok {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}
	if acc.APIKey != "" && ok {
		err = s.GuildMemberRoleAdd(config.DISCORD_GUILD, member.User.ID, config.BIG_BOYE)
		if utils.HandleError(err, "adding role to existing user in agree") {
			_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
			return
		}
	}

	generated, ok := db.CreateUser(member)
	if generated == "" || !ok {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}

	ee, err = s.ChannelMessageEdit(dm.ID, ee.ID, "Your turtl API key is: `"+generated+"`. Please do not lose it or give it to anyone else.\n\nPlease read our getting started guide (<#740011005832069220>) to learn how to install and configure ShareX/turtl.")
	if ee == nil || ee.ID == "" || utils.HandleError(err, "DM in agree") {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}

	err = s.GuildMemberRoleAdd(config.DISCORD_GUILD, member.User.ID, config.BIG_BOYE)
	if utils.HandleError(err, "adding big boye role") {
		_, _ = s.ChannelMessageSend(config.PUB_ALERTS, "<@"+e.UserID+"> An error occurred, please try again later.")
		return
	}
}