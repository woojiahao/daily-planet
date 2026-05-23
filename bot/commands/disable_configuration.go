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

var DisableConfiguration = Command{
	Name:        "disable",
	Group:       "configuration",
	Description: "Disable this source from posting updates",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		disabled := true
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			if _, err := tx.Configuration.UpdateOneByID(
				configurationID,
				models.ConfigurationUpdate{
					Disabled: &disabled,
				},
			); err != nil {
				return err
			}

			if err := context.Scheduler.Cancel(configurationID); err != nil {
				return err
			}

			return nil
		})

		return common.SwitchErrorWithDefaultFunc(
			err,
			helpers.UnknownErrorHandler(),
			map[error]*discordgo.InteractionResponse{
				nil: helpers.CreateSimpleEmbed(
					"Configuration disabled",
					"Configuration for this source has been disabled.\n\nYou will no longer see updates in this source.",
					common.ColorGreen,
				),
				apperrors.ErrConfigurationDBError: helpers.CreateSimpleEmbed(
					"Failed to disable configuration",
					"Failed to disable configuration for this source",
					common.ColorRed,
				),
				apperrors.ErrCronEngineConfigurationNotFound: helpers.CreateSimpleEmbed(
					"Failed to cancel configuration's schedule",
					"Failed to disable configuration's schedule for this source",
					common.ColorRed,
				),
			},
		)
	},
}
