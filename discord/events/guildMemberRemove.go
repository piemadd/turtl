package events

import (
	"github.com/bwmarrin/discordgo"
	"turtl/config"
	"turtl/db"
)

func GuildMemberRemove(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	ok := db.RevokeKey(e.User.ID)
	if !ok {
		_, _ = s.ChannelMessageSend(config.DISCORD_ALERTS, e.Mention()+" left, and I couldn't delete their account.")
	}
}
