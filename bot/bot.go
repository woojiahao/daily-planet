// Package bot contains the implementation for the Discord bot aspects of daily planet.
package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/commands"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db"
)

var database *db.Database

func botToken() string {
	discord_token := os.Getenv("DISCORD_TOKEN")
	if discord_token == "" {
		panic("missing DISCORD_TOKEN in environment")
	}
	return discord_token
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandName := i.ApplicationCommandData().Name
	if handler, ok := commands.CommandHandlerMapping[commandName]; !ok {
		return
	} else {
		callerConfiguration, err := commands.GetCommandCallerConfiguration(i, database)
		if err != nil {
			helpers.SendMessage(s, i, "Failed to execute command. Try again later.")
			return
		}
		context := commands.CommandContext{
			Session:             s,
			Interaction:         i,
			Database:            database,
			CallerConfiguration: callerConfiguration,
		}
		handler(context)
	}
}

func Run() {
	var err error

	database, err = db.New()
	if err != nil {
		panic(err)
	}

	discord, err := discordgo.New("Bot " + botToken())
	if err != nil {
		panic(err)
	}

	discord.AddHandler(interactionHandler)

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	for _, command := range commands.CommandDefinitions {
		_, err = discord.ApplicationCommandCreate(discord.State.User.ID, "1455080799861870673", command)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("bot is running\n")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	discord.Close()
}
