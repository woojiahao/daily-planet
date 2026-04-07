package commands

import (
	"github.com/woojiahao/daily-planet/bot/helpers"
)

func Ping(context CommandContext) {
	helpers.SendMessage(context.Session, context.Interaction, "Pong")
}
