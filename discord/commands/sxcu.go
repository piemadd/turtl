package commands

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/bwmarrin/discordgo"
	"strings"
	"turtl/config"
	"turtl/db"
	"turtl/storage"
	"turtl/utils"
)

func sxcuCommand(s *discordgo.Session, m *discordgo.Message) {
	args := UseArgs(m)

	member, err := s.GuildMember(config.DISCORD_GUILD, m.Author.ID)
	if member == nil || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't find your account. Please DM Polairr to make one.")
		return
	}

	account, ok := db.GetDiscordMemberAccount(member)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	if account.DiscordID == "" || account.APIKey == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I can't find your account. Please DM Polairr to make one.")
		return
	}

	if len(args) < 1 {
		domainString := "```md\n"
		for _, b := range storage.Buckets {
			domainString += "+ " + aws.StringValue(b.Name) + "\n"
		}
		domainString += "```"
		_, _ = s.ChannelMessageSend(m.ChannelID, "**Welcome to turtl!**\n\nWe offer many different domains to choose from. Please pick one below and run `+sxcu <domain of your choice>` to generate a config.\n\nAvailable domains:\n"+domainString+"\n\n**NOTE:** All domains are wildcards. Any character or number, as well as hyphens, can be prepended to the domains. If nothing is prepended, a `i.` will be automatically added.\nExamples: `make-america.great-aga.in`, `cozy.is-stup.id`")
	}

	if !utils.BucketExists(storage.Buckets, args[0]) {
		_, _ = s.ChannelMessageSend(m.ChannelID, "We don't support that domain. Please type `+sxcu` with no arguments to see a list of our domains.")
		return
	}

	if strings.Count(args[0], ".") < 2 {
		args[0] = "i." + args[0]
	}

	dm, err := s.UserChannelCreate(m.Author.ID)
	if dm == nil || err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Please make sure your DMs are open and try running this command again.")
		return
	}

	generatedConfig := strings.NewReader("{\n  \"Version\": \"13.1.0\",\n  \"Name\": \"turtl\",\n  \"DestinationType\": \"ImageUploader, FileUploader\",\n  \"RequestMethod\": \"POST\",\n  \"RequestURL\": \"http://api.turtl.cloud/upload\",\n  \"Body\": \"MultipartFormData\",\n  \"Arguments\": {\n    \"domain\": \"" + args[0] + "\",\n    \"apikey\": \"" + account.APIKey + "\"\n  },\n  \"FileFormName\": \"files[]\",\n  \"URL\": \"$json:files[0].url$\"\n}")
	messageSend := &discordgo.MessageSend{
		Content: "Here is your newly generated ShareX config. Simply download and run it to start using turtl.",
		File: &discordgo.File{
			Name:        "turtl.sxcu",
			ContentType: "text/plain",
			Reader:      generatedConfig,
		},
	}

	_, err = s.ChannelMessageSendComplex(dm.ID, messageSend)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Please make sure your DMs are open and try running this command again.")
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, "Check your DMs!")
}

func init() {
	RegisterCommand(&Command{
		Exec:     sxcuCommand,
		Trigger:  "sxcu",
		Aliases:  []string{"config"},
		Usage:    "sxcu [domain]",
		Desc:     "Generate a ShareX config file",
		Disabled: false,
	})
}
