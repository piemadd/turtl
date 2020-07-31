package commands

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"time"
	"turtl/db"
	"turtl/utils"
)

func purgeCommand(s *discordgo.Session, m *discordgo.Message) {
	allowed, ok := db.CheckAdmin(m)
	if !allowed || !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "You can't use this command, nerd")
		return
	}

	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, HelpManual("purge"))
		return
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "The first argument must be a number!")
		return
	}
	if num > 100 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't delete over 100 messages at once!")
		return
	}

	msg, _ := s.ChannelMessageSend(m.ChannelID, " Deleting "+strconv.Itoa(num)+" messages...")

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if utils.HandleError(err, "sending help message") {
		return
	}

	var toDelete []string
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14).Unix()

	msgs, err := s.ChannelMessages(m.ChannelID, num, "", "", "")
	for _, x := range msgs {
		if x.ID == msg.ID {
			continue
		}
		ts, err := time.Parse(time.RFC3339, string(x.Timestamp))
		if utils.HandleError(err, "sending help message") {
			return
		}
		if ts.Unix() < fourteenDaysAgo {
			continue
		}
		toDelete = append(toDelete, x.ID)
	}
	if utils.HandleError(err, "sending help message") {
		return
	}

	err = s.ChannelMessagesBulkDelete(m.ChannelID, toDelete)
	if utils.HandleError(err, "sending help message") {
		return
	}

	_, _ = s.ChannelMessageEdit(m.ChannelID, msg.ID, "Successfully deleted "+strconv.Itoa(len(toDelete)+1)+" messages")
	time.Sleep(5 * time.Second)
	_ = s.ChannelMessageDelete(m.ChannelID, msg.ID)
}

func init() {
	RegisterCommand(&Command{
		Exec:     purgeCommand,
		Trigger:  "purge",
		Aliases:  nil,
		Usage:    "purge <# of messages to delete>",
		Desc:     "Delete x amount of messages from the channel in which the command was invoked",
		Disabled: false,
	})
}
