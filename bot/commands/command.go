package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
)

type CommandHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate, database *db.Database)

var CommandHandlerMapping = map[string]CommandHandler{
	"ping": Ping,
}

var CommandDefinitions = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Check if the bot is alive",
	},
}
