// Package bot contains the implementation for the Discord bot aspects of daily planet.
package bot

import (
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/commands"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

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

func interactionHandlerWrapper(database *db.Database, commandMap map[commands.CommandIdentifier]commands.Command) interface{} {
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

type BotInterface interface {
	SendMessage(id, message string) error
}

type Scheduler interface {
	Schedule(configurationID models.ConfigurationID) error
	Cancel(configurationID models.ConfigurationID) error
}

type Bot struct {
	session   *discordgo.Session
	scheduler Scheduler
}

func NewBot(database *db.Database) (*Bot, error) {
	session, err := discordgo.New("Bot " + botToken())
	if err != nil {
		return nil, err
	}

	commandsMap := commands.CommandsToNameMap(commands.SupportedCommands)

	session.AddHandler(interactionHandlerWrapper(database, commandsMap))
	return &Bot{session: session}, nil
}

func (b *Bot) SetScheduler(scheduler Scheduler) {
	b.scheduler = scheduler
}

func (b *Bot) Run() error {
	err := b.session.Open()
	if err != nil {
		return err
	}

	existingCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, testGuildID())
	if err != nil {
		return err
	}

	for _, command := range existingCommands {
		err = b.session.ApplicationCommandDelete(
			command.ApplicationID,
			command.GuildID,
			command.ID,
		)
		if err != nil {
			return err
		}
	}

	for _, command := range commands.Commands() {
		_, err = b.session.ApplicationCommandCreate(
			b.session.State.User.ID,
			testGuildID(),
			command,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("bot is running\n")
	return nil
}

func (b *Bot) SendMessage(id, message string) error {
	return nil
}

func (b *Bot) Stop() {
	b.session.Close()
}
