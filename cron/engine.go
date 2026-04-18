// Package cron contains the core Cron scheduling engine that schedules new tasks
package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/woojiahao/daily-planet/bot"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

type CronEngine struct {
	cron     *cron.Cron
	database *db.Database
	bot      bot.BotInterface
	entries  map[models.ConfigurationID]cron.EntryID
}

func NewCronEngine(database *db.Database, bot bot.BotInterface) *CronEngine {
	c := cron.New()
	return &CronEngine{
		cron:     c,
		database: database,
		bot:      bot,
		entries:  make(map[models.ConfigurationID]cron.EntryID),
	}
}

func (ce *CronEngine) Start() error {
	configurations, err := ce.database.Configuration.All()
	if err != nil {
		return err
	}

	for _, configuration := range configurations {
		if configuration.Disabled {
			continue
		}
		err = ce.ScheduleConfiguration(configuration)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ce *CronEngine) Stop() {
	ctx := ce.cron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Second):
	}
}

func (ce *CronEngine) Cancel(configurationID models.ConfigurationID) error {
	if entryID, ok := ce.entries[configurationID]; !ok {
		return fmt.Errorf("configuration not scheduled")
	} else {
		delete(ce.entries, configurationID)
		ce.cron.Remove(entryID)
	}
	return nil
}

func (ce *CronEngine) Schedule(configurationID models.ConfigurationID) error {
	configuration, err := ce.database.Configuration.OneByID(configurationID)
	if err != nil {
		return err
	}
	return ce.ScheduleConfiguration(configuration)
}

func (ce *CronEngine) ScheduleConfiguration(configuration models.Configuration) error {
	entryID, err := ce.cron.AddFunc(configuration.CronSchedule, func() {
		if configuration.Type == models.CommandSourceDM {
			ce.bot.SendMessage(configuration.SnowflakeID, "hi")
		} else {
			if configuration.ChannelID.Valid {
				ce.bot.SendMessage(configuration.ChannelID.String, "hi")
			}
		}
	})
	if err != nil {
		return err
	}

	ce.entries[configuration.ID] = entryID
	return nil
}
