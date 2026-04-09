package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var ListFeeds = Command{
	Name:        "list-feeds",
	Description: "List current feed for the Daily Planet",
	Handler: func(context CommandContext) *discordgo.InteractionResponse {
		feeds, err := context.Database.Feed.All()
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateEmbed(
				"Failed to load feeds",
				"The Daily Planet failed to load feeds for this source",
				helpers.ColorRed,
			)
		}

		var feedURLs []string
		for _, feed := range feeds {
			feedURLs = append(feedURLs, "- "+feed.URL)
		}

		return helpers.CreateEmbed(
			"Feeds fetched",
			strings.Join(feedURLs, "\n"),
			helpers.ColorBlue,
		)
	},
}
