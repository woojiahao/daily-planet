package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var Ping = Command{
	Name:        "ping",
	Description: "Check if the bot is alive",
	Handler: func(context CommandContext) *discordgo.InteractionResponse {
		return helpers.CreateMessage("Pong")
	},
}
