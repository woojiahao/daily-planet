package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	cron_helpers "github.com/woojiahao/daily-planet/cron/helpers"
)

var EditCronSchedule = Command{
	Name:        "edit-cron-schedule",
	Group:       "configuration",
	Description: "Edit the current source's Cron schedule",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "cron-schedule",
			Description: "Cron schedule to replace the current schedule",
			Required:    true,
		},
	},
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		cronSchedule := helpers.GetRequiredOption[string](context, "cron-schedule")
		if !cron_helpers.IsValidCron(cronSchedule) {
			return helpers.CreateSimpleEmbed(
				"Invalid Cron schedule string",
				fmt.Sprintf("Cron schedule string `%s` is not valid.\n\nRefer to [crontab guru](https://crontab.guru) for help creating a Cron schedule string.", cronSchedule),
				helpers.ColorRed,
			)
		}

		err := context.Database.Configuration.UpdateOneByID(context.CallerConfiguration.ID, &cronSchedule, nil, nil, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to update Cron schedule",
				"Failed to update Cron schedule of this source",
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Configuration Cron schedule updated",
			fmt.Sprintf(
				"This source's Cron schedule has been updated from `%s` to `%s`",
				context.CallerConfiguration.CronSchedule,
				cronSchedule,
			),
			helpers.ColorGreen,
		)
	},
}
