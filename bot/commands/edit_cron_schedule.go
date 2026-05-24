package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	cron_helpers "github.com/woojiahao/daily-planet/cron/helpers"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

var EditCronSchedule = Command{Name: "edit-cron-schedule", Group: "configuration", Description: "Edit the current source's Cron schedule", Options: []*discordgo.ApplicationCommandOption{
	{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "cron-schedule",
		Description: "Cron schedule to replace the current schedule",
		Required:    true,
	},
}, Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
	cronSchedule := strings.Trim(helpers.GetRequiredOption[string](context, "cron-schedule"), " ")
	configurationID := context.CallerConfiguration.ID

	err := context.Database.WithTransaction(func(tx db.Database) error {
		if !cron_helpers.IsValidCron(cronSchedule) {
			return apperrors.ErrCronEngineInvalidCronString
		}

		updatedConfiguration, err := tx.Configuration.UpdateOneByID(configurationID, models.ConfigurationUpdate{
			CronSchedule: &cronSchedule,
		})
		if err != nil {
			return err
		}

		if err = context.Scheduler.Cancel(updatedConfiguration.ID); err != nil {
			return err
		}

		if err = context.Scheduler.Schedule(updatedConfiguration); err != nil {
			return err
		}

		return nil
	})

	return common.SwitchErrorWithDefaultFunc(
		err,
		helpers.UnknownErrorHandler(),
		map[error]*discordgo.InteractionResponse{
			nil: helpers.CreateSimpleEmbed(
				"Configuration Cron schedule updated",
				fmt.Sprintf(
					"This source's Cron schedule has been updated from `%s` to `%s`",
					context.CallerConfiguration.CronSchedule,
					cronSchedule,
				),
				common.ColorGreen,
			),
			apperrors.ErrCronEngineInvalidCronString: helpers.CreateSimpleEmbed(
				"Invalid Cron schedule string",
				fmt.Sprintf("Cron schedule string `%s` is not valid or you are using '*' for the seconds field.\n\nUse '0' in the seconds field.\n\nRefer to [crontab guru](https://crontab.guru) for help creating a Cron schedule string.", cronSchedule),
				common.ColorRed,
			),
			apperrors.ErrConfigurationUpdateFailed: helpers.CreateSimpleEmbed(
				"Failed to update Cron schedule",
				"Failed to update Cron schedule of this source",
				common.ColorRed,
			),
			apperrors.ErrCronEngineConfigurationNotFound: helpers.CreateSimpleEmbed(
				"Failed to cancel configuration's schedule",
				"Failed to disable configuration's schedule for this source",
				common.ColorRed,
			),
			apperrors.ErrCronEngineScheduleAlreadyRunning: helpers.CreateSimpleEmbed(
				"Failed to start configuration's schedule",
				"Failed to start configuration's schedule for this source as the cron is currently already running",
				common.ColorRed,
			),
			apperrors.ErrCronEngineScheduleError: helpers.CreateSimpleEmbed(
				"Failed to start configuration's schedule",
				"Failed to start configuration's schedule for this source",
				common.ColorRed,
			),
		},
	)
}}
