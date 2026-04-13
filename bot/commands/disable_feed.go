package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db/models"
)

var DisableFeed = Command{
	Name:        "disable",
	Group:       "feed",
	Description: "Disable a provided feed",
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
		feed, err := context.Database.Feed.OneByKey(models.NewFeedKey(context.CallerConfiguration.ID, url))
		if err != nil {
			if err == sql.ErrNoRows {
				return helpers.CreateSimpleEmbed(
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse `/list-feeds` to verify that it exists in this source.", url),
					helpers.ColorRed,
				)
			}
			return helpers.CreateSimpleEmbed(
				"Failed to fetch feed",
				fmt.Sprintf("Failed to fetch feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		err = context.Database.Feed.UpdateOneByID(feed.ID, true)
		if err != nil {
			return helpers.CreateSimpleEmbed(
				"Failed to disable feed",
				fmt.Sprintf("Failed to disable feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feed disabled",
			fmt.Sprintf("Feed %s disabled", url),
			helpers.ColorGreen,
		)
	},
}
