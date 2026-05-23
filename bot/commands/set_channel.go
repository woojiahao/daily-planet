package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
)

var SetChannel = Command{
	Name:        "set-channel",
	Description: "Set the configuration channel for servers to the channel this command is invoked in",
	Group:       CommandGroupConfiguration,
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		if context.Interaction.GuildID == "" {
			return helpers.CreateSimpleEmbed("Channel not required", "You do not need to set the channel in a DM. We will send updates directly to your DM.", common.ColorBlue)
		}

		currentChannelID := context.Interaction.ChannelID
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			_, err := tx.Configuration.UpdateOneByID(
				configurationID,
				models.ConfigurationUpdate{
					ChannelID: &currentChannelID,
				},
			)
			return err
		})

		return common.SwitchErrorWithDefaultFunc(err, helpers.UnknownErrorHandler(), map[error]*discordgo.InteractionResponse{
			nil: helpers.CreateSimpleEmbed(
				"Channel set",
				fmt.Sprintf("Channel for this source set to <#%s>", currentChannelID),
				common.ColorGreen,
			),
			apperrors.ErrConfigurationUpdateFailed: helpers.CreateSimpleEmbed(
				"Failed to set channel",
				"Failed to set channel for this source.",
				common.ColorRed,
			),
		})
	},
}
