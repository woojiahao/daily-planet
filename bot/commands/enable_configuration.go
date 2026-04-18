package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var EnableConfiguration = Command{
	Name:        "enable",
	Group:       "configuration",
	Description: "Enable this source to allow updates to be posted",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		disabled := false
		_, err := context.Database.Configuration.UpdateOneByID(context.CallerConfiguration.ID, nil, nil, nil, &disabled)
		if err != nil {
			fmt.Printf("%v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to enable configuration",
				"Failed to enable configuration for this source",
				helpers.ColorRed,
			)
		}

		if err = context.Scheduler.Schedule(context.CallerConfiguration); err != nil {
			fmt.Printf("%v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to start configuration's schedule",
				"Failed to start configuration's schedule for this source",
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Configuration enabled",
			"Configuration for this source has been enabled.\n\nYou will start receiving updates in this source.",
			helpers.ColorGreen,
		)
	},
}
