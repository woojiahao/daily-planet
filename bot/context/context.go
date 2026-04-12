package context

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

type CommandContext struct {
	Session             *discordgo.Session
	Interaction         *discordgo.InteractionCreate
	Database            *db.Database
	CallerConfiguration models.Configuration
}
