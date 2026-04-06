package commands

import "github.com/bwmarrin/discordgo"

type CommandHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate)

var CommandHandlerMapping = map[string]CommandHandler{
	"ping": Ping,
}

var CommandDefinitions = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Check if the bot is alive",
	},
}
