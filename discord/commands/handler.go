package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

var CommandMap = make(map[string]*Command)
var AliasMap = make(map[string]string)

type Command struct {
	Exec       func(*discordgo.Session, *discordgo.Message)
	Trigger    string
	Aliases    []string
	Usage      string
	Desc       string
	Disabled   bool
}

func RegisterCommand(c *Command) {
	CommandMap[c.Trigger] = c
	for _, alias := range c.Aliases {
		AliasMap[alias] = c.Trigger
	}

	log.Printf("%s loaded!", c.Trigger)
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil {
		return
	}
	if m.Author == nil {
		return
	}
	if m.Author.Bot {
		return
	}
	if m.Author.ID == "" {
		return
	}
	if m.Message == nil {
		return
	}
	if m.Message.Content == "" {
		return
	}
	if m.Member == nil {
		return
	}

	// command logic
	prefix := "+"
	if len(m.Message.Content) <= 1 || m.Message.Content[0:len(prefix)] != prefix {
		return
	}
	cmdTrigger := strings.Split(m.Content, " ")[0][len(prefix):]
	cmdTrigger = strings.ToLower(cmdTrigger)
	cmd, ok := CommandMap[cmdTrigger]
	if !ok {
		cmd, ok = CommandMap[AliasMap[cmdTrigger]]
		if !ok {
			return
		}
	}

	//if cmd.Disabled && m.Author.ID != constants.DEV_ID {
	//	return
	//}

	// perms
	//if strings.ToLower(cmd.Category) == "developer" && m.Author.ID != constants.DEV_ID {
	//	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, utils.ErrorEmbed("You do not have permission to execute that command!", m.Message))
	//	return
	//}

	// if all is good, execute
	cmd.Exec(s, m.Message)
}