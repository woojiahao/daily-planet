package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/ds"
)

func GetCommandCallerConfiguration(
	interaction *discordgo.InteractionCreate,
	scheduler context.Scheduler,
	database *db.Database,
) (models.Configuration, error) {
	var currentCommandSource models.CommandSource
	var snowflakeID string
	if interaction.GuildID == "" {
		currentCommandSource = models.CommandSourceDM
		snowflakeID = interaction.User.ID
	} else {
		currentCommandSource = models.CommandSourceServer
		snowflakeID = interaction.GuildID
	}

	configuration, err := database.Configuration.OneBySnowflakeID(snowflakeID)
	fmt.Printf("configuration is %v and err is %v, and snowflakeID is %v and %v\n", configuration, err, snowflakeID, err == sql.ErrNoRows)
	if err != nil {
		if err == sql.ErrNoRows {
			// no configuration belonging to the current command source, create one
			channelID := &interaction.ChannelID
			iErr := database.Configuration.InsertOne(snowflakeID, channelID, currentCommandSource)
			if iErr != nil {
				fmt.Printf("iErr is %v\n", iErr)
				return models.Configuration{}, iErr
			}

			configuration, iErr := database.Configuration.OneBySnowflakeID(snowflakeID)
			if iErr != nil {
				fmt.Printf("iErr is %v\n", iErr)
				return models.Configuration{}, iErr
			}
			iErr = scheduler.Schedule(configuration)
			if iErr != nil {
				fmt.Printf("iErr is %v\n", iErr)
				return models.Configuration{}, iErr
			}

			return configuration, nil
		} else {
			return models.Configuration{}, err
		}
	}

	return configuration, nil
}

type (
	CommandHandler            func(context context.CommandContext) *discordgo.InteractionResponse
	CommandModalSubmitHandler func(context context.CommandContext) *discordgo.InteractionResponse
)

type (
	CommandName       string
	CommandGroup      string
	CommandIdentifier ds.ComparablePair[CommandGroup, CommandName]
)

const (
	CommandGroupFeed          CommandGroup = "feed"
	CommandGroupConfiguration CommandGroup = "configuration"
)

func NewCommandIdentifier(group, name string) CommandIdentifier {
	return CommandIdentifier(*ds.NewComparablePair(CommandGroup(group), CommandName(name)))
}

// TODO(woojiahao): the definition should actually just expand to avoid nesting structs unnecessarily
type CommandDefinition struct {
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

type Command struct {
	Name               CommandName
	Group              CommandGroup
	Description        string
	Options            []*discordgo.ApplicationCommandOption
	Handler            CommandHandler
	ModalSubmitHandler CommandModalSubmitHandler
}

func (c Command) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        string(c.Name),
		Description: c.Description,
		Options:     c.Options,
	}
}

func (c Command) Identifier() CommandIdentifier {
	return CommandIdentifier(*ds.NewComparablePair(c.Group, c.Name))
}

var groupDescriptions = map[CommandGroup]string{
	CommandGroupFeed:          "Modify the feeds that are maintained in this source",
	CommandGroupConfiguration: "Modify the configuration of this source",
}

var SupportedCommands = []Command{
	// no group
	Ping,

	// feed related
	ListFeeds,
	AddFeed,
	DeleteFeed,
	DisableFeed,
	EnableFeed,
	FetchFeed,
	FetchFeeds,

	// configuration related
	EditCronSchedule,
	DisableConfiguration,
	EnableConfiguration,
	SetChannel,
}

func Commands() []*discordgo.ApplicationCommand {
	var allCommands []*discordgo.ApplicationCommand
	groupedCommands := make(map[CommandGroup][]*discordgo.ApplicationCommand)

	for _, command := range SupportedCommands {
		if command.Group == "" {
			allCommands = append(allCommands, command.ToDiscordCommand())
		} else {
			groupedCommands[command.Group] = append(groupedCommands[command.Group], command.ToDiscordCommand())
		}
	}

	for group, commands := range groupedCommands {
		var subCommands []*discordgo.ApplicationCommandOption
		for _, command := range commands {
			subCommands = append(subCommands, &discordgo.ApplicationCommandOption{
				Name:        command.Name,
				Description: command.Description,
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options:     command.Options,
			})
		}
		allCommands = append(allCommands, &discordgo.ApplicationCommand{
			Name:        string(group),
			Description: groupDescriptions[group],
			Options:     subCommands,
		})
	}

	return allCommands
}

func CommandsToNameMap(commands []Command) map[CommandIdentifier]Command {
	commandByName := make(map[CommandIdentifier]Command)

	for _, command := range commands {
		commandByName[command.Identifier()] = command
	}

	return commandByName
}
