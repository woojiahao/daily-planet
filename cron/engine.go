// Package cron contains the core Cron scheduling engine that schedules new tasks
package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/woojiahao/daily-planet/bot"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

type CronEngine struct {
	cron     *cron.Cron
	database *db.Database
	bot      *bot.BotInterface
	entries  map[models.ConfigurationID]cron.EntryID
}

func NewCronEngine(database *db.Database, bot bot.BotInterface) *CronEngine {
	c := cron.New()
	return &CronEngine{
		cron:     c,
		database: database,
		bot:      &bot,
		entries:  make(map[models.ConfigurationID]cron.EntryID),
	}
}

func (ce *CronEngine) Start() {
}

func (ce *CronEngine) CancelConfiguration(configurationID models.ConfigurationID) error {
	return nil
}

func (ce *CronEngine) ScheduleConfiguration(configurationID models.ConfigurationID) error {
	configuration, err := ce.database.Configuration.OneByID(configurationID)
	if err != nil {
		return err
	}

	ce.cron.AddFunc(configuration.CronSchedule, func() {
		if configuration.Type == models.CommandSourceDM {
			(*ce.bot).SendMessage(configuration.SnowflakeID, "hi")
		} else {
			if configuration.ChannelID.Valid {
				(*ce.bot).SendMessage(configuration.ChannelID.String, "hi")
			}
		}
	})
	return nil
}
