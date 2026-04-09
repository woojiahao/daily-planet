package commands

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

func GetCommandCallerConfiguration(interaction *discordgo.InteractionCreate, database *db.Database) (models.Configuration, error) {
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
	if err != nil {
		if err == sql.ErrNoRows {
			// no configuration belonging to the current command source, create one
			database.Configuration.InsertOne(snowflakeID, currentCommandSource)
			configuration, err := database.Configuration.OneBySnowflakeID(snowflakeID)
			if err != nil {
				return models.Configuration{}, err
			}

			return configuration, nil
		} else {
			return models.Configuration{}, err
		}
	}

	return configuration, nil
}

type CommandContext struct {
	Session             *discordgo.Session
	Interaction         *discordgo.InteractionCreate
	Database            *db.Database
	CallerConfiguration models.Configuration
}

type CommandHandler func(context CommandContext) *discordgo.InteractionResponse

type CommandName string

// TODO(woojiahao): the definition should actually just expand to avoid nesting structs unnecessarily
type CommandDefinition struct {
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

type Command struct {
	Name        CommandName
	Description string
	Options     []*discordgo.ApplicationCommandOption
	Handler     CommandHandler
}

func (c Command) ToDiscordCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        string(c.Name),
		Description: c.Description,
		Options:     c.Options,
	}
}

var SupportedCommands = []Command{
	Ping,

	// feed related
	ListFeeds,
	AddFeed,
	DeleteFeed,
	EditFeed,
}

func CommandsToNameMap(commands []Command) map[CommandName]Command {
	commandByName := make(map[CommandName]Command)

	for _, command := range commands {
		commandByName[command.Name] = command
	}

	return commandByName
}
