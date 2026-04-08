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

type CommandHandler func(context CommandContext)

var CommandHandlerMapping = map[string]CommandHandler{
	"ping":      Ping,
	"list-feed": ListFeeds,
	"add-feed":  AddFeed,
}

var CommandDefinitions = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Check if the bot is alive",
	},
	{
		Name:        "list-feed",
		Description: "List current feed for the Daily Planet",
	},
	{
		Name:        "add-feed",
		Description: "Add a feed to the Daily Planet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "Feed URL to add",
				Required:    true,
			},
		},
	},
}
