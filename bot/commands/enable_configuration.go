package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

var EnableConfiguration = Command{
	Name:        "enable",
	Group:       "configuration",
	Description: "Enable this source to allow updates to be posted",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		disabled := false
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			_, err := tx.Configuration.UpdateOneByID(configurationID, models.ConfigurationUpdate{
				Disabled: &disabled,
			})
			if err != nil {
				return err
			}

			if err = context.Scheduler.Schedule(context.CallerConfiguration); err != nil {
				return err
			}

			return nil
		})

		return common.SwitchErrorWithDefaultFunc(err, helpers.UnknownErrorHandler(), map[error]*discordgo.InteractionResponse{
			nil: helpers.CreateSimpleEmbed(
				"Configuration enabled",
				"Configuration for this source has been enabled.\n\nYou will start receiving updates in this source.",
				common.ColorGreen,
			),
			apperrors.ErrConfigurationUpdateFailed: helpers.CreateSimpleEmbed(
				"Failed to enable configuration",
				"Failed to enable configuration for this source",
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
		})
	},
}
