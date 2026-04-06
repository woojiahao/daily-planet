package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
	helpers "github.com/woojiahao/daily-planet/db/utility"
)

func Ping(s *discordgo.Session, i *discordgo.InteractionCreate, database *db.Database) {
	helpers.SendMessage(s, i, "Pong")
}
