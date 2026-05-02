// Package cron contains the core Cron scheduling engine that schedules new tasks
package cron

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/woojiahao/daily-planet/bot"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/source"
)

type CronEngine struct {
	cron     *cron.Cron
	database *db.Database
	bot      bot.BotInterface
	entries  map[models.ConfigurationID]cron.EntryID
}

func NewCronEngine(database *db.Database, bot bot.BotInterface) *CronEngine {
	c := cron.New(cron.WithSeconds())
	return &CronEngine{
		cron:     c,
		database: database,
		bot:      bot,
		entries:  make(map[models.ConfigurationID]cron.EntryID),
	}
}

func (ce *CronEngine) Start() error {
	ce.cron.Start()

	configurations, err := ce.database.Configuration.All()
	if err != nil {
		return err
	}

	for _, configuration := range configurations {
		if configuration.Disabled {
			continue
		}
		err = ce.Schedule(configuration)
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
		ce.cron.Remove(entryID)
		delete(ce.entries, configurationID)
	}
	return nil
}

func (ce *CronEngine) ScheduleConfiguration(configurationID models.ConfigurationID) error {
	configuration, err := ce.database.Configuration.OneByID(configurationID)
	if err != nil {
		return err
	}
	return ce.Schedule(configuration)
}

func (ce *CronEngine) Schedule(configuration models.Configuration) error {
	if _, ok := ce.entries[configuration.ID]; ok {
		return fmt.Errorf("schedule for configuration is already running. call Cancel on it first")
	}

	fmt.Printf("scheduled configuration %d with schedule %s\n", configuration.ID, configuration.CronSchedule)

	entryID, err := ce.cron.AddFunc(
		configuration.CronSchedule,
		func() {
			if !configuration.ChannelID.Valid {
				fmt.Printf("no channel ID configured for %s\n", configuration.SnowflakeID)
				return
			}

			configurationID := configuration.ID
			channelID := configuration.ChannelID.String
			fmt.Printf("trigger with type: %s and channel id %v\n", configuration.Type, configuration.ChannelID)

			source.FetchFeedsAlgorithmWrapper(
				configurationID,
				ce.database,
				false,
				func(title string, description string, color common.Color) {
					err := ce.bot.SendSimpleEmbed(
						channelID,
						title,
						description,
						color,
					)
					if err != nil {
						fmt.Printf("err is %v\n", err)
					}
				},
			)
		},
	)
	if err != nil {
		return err
	}

	ce.entries[configuration.ID] = entryID
	return nil
}
