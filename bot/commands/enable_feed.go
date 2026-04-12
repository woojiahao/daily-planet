package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var EnableFeed = Command{
	Name:        "enable-feed",
	Description: "Enable a provided feed",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "Feed URL to add",
			Required:    true,
		},
	},
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		url := helpers.GetRequiredOption[string](context, "url")
		feed, err := context.Database.Feed.OneByConfigurationIDAndURL(context.CallerConfiguration.ID, url)
		if err != nil {
			if err == sql.ErrNoRows {
				return helpers.CreateSimpleEmbed(
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse /list-feeds to verify that it exists in this source.", url),
					helpers.ColorRed,
				)
			}
			return helpers.CreateSimpleEmbed(
				"Failed to fetch feed",
				fmt.Sprintf("Failed to fetch feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		err = context.Database.Feed.UpdateOneByID(feed.ID, false)
		if err != nil {
			return helpers.CreateSimpleEmbed(
				"Failed to enable feed",
				fmt.Sprintf("Failed to enable feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feed enabled",
			fmt.Sprintf("Feed %s enabled", url),
			helpers.ColorGreen,
		)
	},
}
