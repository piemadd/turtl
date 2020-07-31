package events

import (
	"github.com/bwmarrin/discordgo"
)

func GuildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	dm, _ := s.UserChannelCreate(e.User.ID)
	_, _ = s.ChannelMessageSend(dm.ID, "Welcome to turtl! Please read our guidelines and privacy notice in <#737854426693369899>. If you agree to our guidelines, please DM Polairr#0001 to get an account.")
}
