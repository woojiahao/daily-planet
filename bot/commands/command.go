package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
)

type commandSource string

const (
	commandSourceDM     commandSource = "DM"
	commandSourceServer               = "server"
)

func getCommandSource(interaction *discordgo.InteractionCreate) commandSource {
	if interaction.GuildID == "" {
		return commandSourceDM
	}

	return commandSourceServer
}

type CommandHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate, database *db.Database)

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
