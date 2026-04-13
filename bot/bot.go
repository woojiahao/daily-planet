// Package bot contains the implementation for the Discord bot aspects of daily planet.
package bot

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/commands"
	"github.com/woojiahao/daily-planet/bot/context"
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

func interactionHandlerWrapper(commandMap map[commands.CommandIdentifier]commands.Command) interface{} {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		callerConfiguration, err := commands.GetCommandCallerConfiguration(interaction, database)
		context := context.CommandContext{
			Session:             session,
			Interaction:         interaction,
			Database:            database,
			CallerConfiguration: callerConfiguration,
		}

		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			data := interaction.ApplicationCommandData()
			commandName := data.Name
			var commandSub string

			for _, opt := range data.Options {
				if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
					commandSub = opt.Name
				}
			}
			if command, ok := commandMap[commands.NewCommandIdentifier(commandName, commandSub)]; ok {
				if err != nil {
					helpers.SendMessage(session, interaction, "Failed to execute command. Try again later.")
					return
				}
				response := command.Handler(context)
				if response != nil {
					// printing response, else assume that the handler did something already
					err = session.InteractionRespond(interaction.Interaction, response)
					if err != nil {
						// TODO(woojiahao): maybe print something
						fmt.Printf("err is %v\n", err)
					}
				}
			}

		case discordgo.InteractionModalSubmit:
			data := context.Interaction.ModalSubmitData()
			parts := strings.Split(data.CustomID, ":")
			commandName := parts[0]
			commandSub := parts[1]
			if command, ok := commandMap[commands.NewCommandIdentifier(commandName, commandSub)]; ok {
				if command.ModalSubmitHandler != nil {
					response := command.ModalSubmitHandler(context)
					if response != nil {
						err = session.InteractionRespond(interaction.Interaction, response)
						if err != nil {
							// TODO(woojiahao): maybe print something
							fmt.Printf("err is %v\n", err)
						}
					}
				}
			}

		default:
			return
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

	for _, command := range commands.Commands() {
		_, err = discord.ApplicationCommandCreate(
			discord.State.User.ID,
			testGuildID(),
			command,
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
