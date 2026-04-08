package commands

import (
	"fmt"
	"strings"

	"github.com/woojiahao/daily-planet/bot/helpers"
)

var ListFeeds Command = Command{
	Name:        "list-feeds",
	Description: "List current feed for the Daily Planet",
	Handler: func(context CommandContext) {
		feeds, err := context.Database.Feed.All()
		if err != nil {
			fmt.Printf("err is %v\n", err)
			helpers.SendEmbed(
				context.Session,
				context.Interaction,
				"Failed to load feeds",
				"The Daily Planet failed to load feeds for this source",
				helpers.ColorRed,
			)
			return
		}

		var feedURLs []string
		for _, feed := range feeds {
			feedURLs = append(feedURLs, "- "+feed.URL)
		}

		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Feeds fetched",
			strings.Join(feedURLs, "\n"),
			helpers.ColorBlue,
		)
	},
}
