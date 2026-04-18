package context

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

type Scheduler interface {
	Schedule(configuration models.Configuration) error
	Cancel(configurationID models.ConfigurationID) error
}

type CommandContext struct {
	Session             *discordgo.Session
	Interaction         *discordgo.InteractionCreate
	Database            *db.Database
	Scheduler           Scheduler
	CallerConfiguration models.Configuration
}
