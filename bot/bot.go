// Package bot contains the implementation for the Discord bot aspects of daily planet.
package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/commands"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
)

type BotInterface interface {
	SendMessage(id, content string) error
	SendSimpleEmbed(id, title, description string, color common.Color) error
}

type Bot struct {
	database  *db.Database
	session   *discordgo.Session
	scheduler context.Scheduler
}

func interactionHandlerWrapper(database *db.Database, scheduler context.Scheduler, commandMap map[commands.CommandIdentifier]commands.Command) interface{} {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		callerConfiguration, err := commands.GetCommandCallerConfiguration(interaction, database)
		context := context.CommandContext{
			Session:             session,
			Interaction:         interaction,
			Database:            database,
			Scheduler:           scheduler,
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

func NewBot(database *db.Database) (*Bot, error) {
	session, err := discordgo.New("Bot " + helpers.BotToken())
	if err != nil {
		return nil, err
	}

	return &Bot{database: database, session: session}, nil
}

func (b *Bot) SetScheduler(scheduler context.Scheduler) {
	b.scheduler = scheduler
}

func (b *Bot) Run() error {
	commandsMap := commands.CommandsToNameMap(commands.SupportedCommands)

	b.session.AddHandler(interactionHandlerWrapper(b.database, b.scheduler, commandsMap))

	err := b.session.Open()
	if err != nil {
		return err
	}

	existingCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, helpers.TestGuildID())
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
			helpers.TestGuildID(),
			command,
		)
		if err != nil {
			return err
		}
	}

	fmt.Printf("bot is running\n")
	return nil
}

func (b *Bot) SendMessage(id, content string) error {
	_, err := b.session.ChannelMessageSend(id, content)
	return err
}

func (b *Bot) SendSimpleEmbed(id, title, description string, color common.Color) error {
	_, err := b.session.ChannelMessageSendEmbed(id, &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       int(color),
	})
	return err
}

func (b *Bot) Stop() {
	b.session.Close()
}
