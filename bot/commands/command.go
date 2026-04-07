package commands

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

func GetCommandSource(interaction *discordgo.InteractionCreate, database *db.Database) (models.CommandSource, error) {
	var currentCommandSource models.CommandSource
	var snowflakeID string
	if interaction.GuildID == "" {
		currentCommandSource = models.CommandSourceDM
		snowflakeID = interaction.User.ID
	} else {
		currentCommandSource = models.CommandSourceServer
		snowflakeID = interaction.GuildID
	}

	// TODO(woojiahao): maybe adopt HasOne instead?
	_, err := database.Configuration.OneBySnowflakeID(snowflakeID)
	if err != nil {
		if err == sql.ErrNoRows {
			// no configuration belonging to the current command source, create one
			database.Configuration.Insert(snowflakeID, currentCommandSource)
		} else {
			return "", err
		}
	}

	return currentCommandSource, nil
}

type CommandContext struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Database    *db.Database
	Source      models.CommandSource
}

type CommandHandler func(context CommandContext)

var CommandHandlerMapping = map[string]CommandHandler{
	"ping":      Ping,
	"list-feed": ListFeed,
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
}
