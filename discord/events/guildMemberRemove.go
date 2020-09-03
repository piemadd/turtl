package events

import (
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"turtl/db"
)

func GuildMemberRemove(s *discordgo.Session, e *discordgo.GuildMemberRemove) {
	ok := db.RevokeKey(e.User.ID)
	if !ok {
		_, _ = s.ChannelMessageSend(os.Getenv("DISCORD_ALERTS"), e.Mention()+" left, and I couldn't delete their account.")
	}
}
