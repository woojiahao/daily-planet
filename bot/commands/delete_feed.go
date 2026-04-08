package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var DeleteFeed = Command{
	Name:        "delete-feed",
	Description: "Deletes a feed from the Daily Planet",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "Feed URL to delete",
			Required:    true,
		},
	},
	Handler: func(context CommandContext) {
		// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
		url := strings.Trim(context.Interaction.ApplicationCommandData().Options[0].StringValue(), " ")

		feed, err := context.Database.Feed.OneByConfigurationIDAndURL(context.CallerConfiguration.ID, url)
		if err != nil {
			if err == sql.ErrNoRows {
				helpers.SendEmbed(
					context.Session,
					context.Interaction,
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse /list-feeds to verify that it exists in this source.", url),
					helpers.ColorRed,
				)
			} else {
				helpers.SendEmbed(
					context.Session,
					context.Interaction,
					"Failed to fetch feed",
					fmt.Sprintf("Failed to fetch feed by URL %s. Try again", url),
					helpers.ColorRed,
				)
			}
			return
		}

		err = context.Database.Feed.DeleteOneByID(feed.ID)
		if err != nil {
			helpers.SendEmbed(
				context.Session,
				context.Interaction,
				"Failed to delete feed",
				fmt.Sprintf("Failed to delete feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
			return
		}

		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Feed deleted",
			fmt.Sprintf("Feed %s has been deleted from this source", url),
			helpers.ColorGreen,
		)
	},
}
