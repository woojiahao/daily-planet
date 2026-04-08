package commands

import (
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var Ping Command = Command{
	Name:        "ping",
	Description: "Check if the bot is alive",
	Handler: func(context CommandContext) {
		helpers.SendMessage(context.Session, context.Interaction, "Pong")
	},
}
