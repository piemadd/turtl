package events

import (
	"github.com/bwmarrin/discordgo"
)

func GuildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	dm, _ := s.UserChannelCreate(e.User.ID)
	_, _ = s.ChannelMessageSend(dm.ID, "Welcome to turtl! Please read our guidelines and privacy notice in <#737854426693369899>. After you've read them, please react with the checkmark to show that you agree to our guidelines and privacy notice. After this, I will DM you with your very own API key and instructions on how to get started.")
}
