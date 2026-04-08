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

type interactionHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate)

func interactionHandlerWrapper(commandMap map[commands.CommandName]commands.Command) interactionHandler {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.Type != discordgo.InteractionApplicationCommand {
			return
		}

		commandName := interaction.ApplicationCommandData().Name
		if command, ok := commandMap[commands.CommandName(commandName)]; !ok {
			return
		} else {
			callerConfiguration, err := commands.GetCommandCallerConfiguration(interaction, database)
			if err != nil {
				helpers.SendMessage(session, interaction, "Failed to execute command. Try again later.")
				return
			}
			context := commands.CommandContext{
				Session:             session,
				Interaction:         interaction,
				Database:            database,
				CallerConfiguration: callerConfiguration,
			}
			command.Handler(context)
		}
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

	commandsMap := commands.CommandsToNameMap(commands.SupportedCommands)

	discord.AddHandler(interactionHandlerWrapper(commandsMap))

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	for _, command := range commands.SupportedCommands {
		_, err = discord.ApplicationCommandCreate(
			discord.State.User.ID,
			// TODO(woojiahao): set this as a configurable env var so that we don't hard code it in the commits
			// no sensitive information, just to avoid hardcoding
			"1455080799861870673",
			command.ToDiscordCommand(),
		)
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
