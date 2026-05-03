package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
)

var (
	errDisableConfigurationUpdateFailed         = errors.New("failed to update configuration row")
	errDisableConfigurationCancelScheduleFailed = errors.New("failed to cancel configuration schedule")
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
				nil,
				nil,
				nil,
				&disabled,
			); err != nil {
				return errDisableConfigurationUpdateFailed
			}

			if err := context.Scheduler.Cancel(configurationID); err != nil {
				return errDisableConfigurationCancelScheduleFailed
			}

			return nil
		})

		return common.SwitchError(err, map[error]*discordgo.InteractionResponse{
			nil: helpers.CreateSimpleEmbed(
				"Configuration disabled",
				"Configuration for this source has been disabled.\n\nYou will no longer see updates in this source.",
				common.ColorGreen,
			),
			errDisableConfigurationUpdateFailed: helpers.CreateSimpleEmbed(
				"Failed to disable configuration",
				"Failed to disable configuration for this source",
				common.ColorRed,
			),
			errDisableConfigurationCancelScheduleFailed: helpers.CreateSimpleEmbed(
				"Failed to cancel configuration's schedule",
				"Failed to disable configuration's schedule for this source",
				common.ColorRed,
			),
		})
	},
}
