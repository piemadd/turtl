package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func UseArgs(m *discordgo.Message) []string {
	args := strings.Split(m.Content, " ")
	copy(args[0:], args[1:])
	args[len(args)-1] = ""
	args = args[:len(args)-1]

	return args
}
