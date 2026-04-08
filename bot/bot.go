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
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		panic("missing DISCORD_TOKEN in environment")
	}
	return discordToken
}

func testGuildID() string {
	return os.Getenv("TEST_GUILD_ID")
}

func interactionHandlerWrapper(commandMap map[commands.CommandName]commands.Command) interface{} {
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

	existingCommands, err := discord.ApplicationCommands(discord.State.User.ID, testGuildID())
	if err != nil {
		panic(err)
	}

	for _, command := range existingCommands {
		err = discord.ApplicationCommandDelete(
			command.ApplicationID,
			command.GuildID,
			command.ID,
		)
		if err != nil {
			panic(err)
		}
	}

	for _, command := range commands.SupportedCommands {
		_, err = discord.ApplicationCommandCreate(
			discord.State.User.ID,
			testGuildID(),
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
