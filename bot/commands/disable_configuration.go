package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var DisableConfiguration = Command{
	Name:        "disable",
	Group:       "configuration",
	Description: "Disable this source from posting updates",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		disabled := true
		_, err := context.Database.Configuration.UpdateOneByID(context.CallerConfiguration.ID, nil, nil, nil, &disabled)
		if err != nil {
			fmt.Printf("%v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to disable configuration",
				"Failed to disable configuration for this source",
				helpers.ColorRed,
			)
		}

		if err = context.Scheduler.Cancel(context.CallerConfiguration.ID); err != nil {
			fmt.Printf("%v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to cancel configuration's schedule",
				"Failed to disable configuration's schedule for this source",
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Configuration disabled",
			"Configuration for this source has been disabled.\n\nYou will no longer see updates in this source.",
			helpers.ColorGreen,
		)
	},
}
