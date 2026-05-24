// Package cron contains the core Cron scheduling engine that schedules new tasks
package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/woojiahao/daily-planet/apperrors"
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

	running sync.Map // ConfigurationID -> struct{}
}

func NewCronEngine(database *db.Database, bot bot.BotInterface) *CronEngine {
	c := cron.New(cron.WithSeconds())
	return &CronEngine{
		cron:     c,
		database: database,
		bot:      bot,
		entries:  make(map[models.ConfigurationID]cron.EntryID),
		running:  sync.Map{},
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
		return apperrors.ErrCronEngineConfigurationNotFound
	} else {
		fmt.Printf("cancelling schedule for configuration %d with entryID %d\n", configurationID, entryID)
		ce.running.Delete(configurationID)
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
		return apperrors.ErrCronEngineScheduleAlreadyRunning
	}

	fmt.Printf(
		"scheduled configuration %d with schedule %s at time %d\n",
		configuration.ID,
		configuration.CronSchedule,
		time.Now().UnixMilli(),
	)

	entryID, err := ce.cron.AddFunc(
		configuration.CronSchedule,
		func() {
			if _, loaded := ce.running.LoadOrStore(configuration.ID, struct{}{}); loaded {
				fmt.Printf("skipping run for config %d (already running)\n", configuration.ID)
				return
			}

			defer ce.running.Delete(configuration.ID)

			if !configuration.ChannelID.Valid {
				fmt.Printf("no channel ID configured for %s\n", configuration.SnowflakeID)
				return
			}

			configurationID := configuration.ID
			channelID := configuration.ChannelID.String
			fmt.Printf(
				"trigger with type: %s and channel id %v at time %d\n",
				configuration.Type,
				configuration.ChannelID,
				time.Now().UnixMilli(),
			)

			ce.database.WithTransaction(func(tx db.Database) error {
				source.FetchFeedsAlgorithmWrapper(
					configurationID,
					&tx,
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
				return nil
			})
		},
	)
	if err != nil {
		return common.WrapError(apperrors.ErrCronEngineScheduleError, err)
	}

	ce.entries[configuration.ID] = entryID
	return nil
}
