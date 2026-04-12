package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db/models"
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
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
		url := strings.Trim(context.Interaction.ApplicationCommandData().Options[0].StringValue(), " ")

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

		err = context.Database.Feed.DeleteOneByID(feed.ID)
		if err != nil {
			return helpers.CreateSimpleEmbed(
				"Failed to delete feed",
				fmt.Sprintf("Failed to delete feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feed deleted",
			fmt.Sprintf("Feed %s has been deleted from this source", url),
			helpers.ColorGreen,
		)
	},
}
