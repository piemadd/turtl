package discord

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"turtl/discord/commands"
	"turtl/discord/events"

	_ "github.com/joho/godotenv/autoload"
)

var Client *discordgo.Session

func CreateBot() {
	var err error
	Client, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}

	Client.AddHandler(events.Agree)
	Client.AddHandler(events.GuildMemberRemove)
	Client.AddHandler(events.GuildMemberAdd)
	Client.AddHandler(commands.HandleCommand)

	err = Client.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}
}
