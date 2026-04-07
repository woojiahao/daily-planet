package commands

import (
	"fmt"

	"github.com/woojiahao/daily-planet/bot/helpers"
)

func ListFeeds(context CommandContext) {
	feeds, err := context.Database.Feed.All()
	if err != nil {
		fmt.Printf("underlying err is %v\n", err)
		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Failed to load feeds",
			"The Daily Planet failed to load feeds for this source.",
			0x5865F2,
		)
		return
	}

	helpers.SendEmbed(
		context.Session,
		context.Interaction,
		"Feeds fetched",
		fmt.Sprintf("These are the feeds for the current source: %v", feeds),
		0x5865F2,
	)
}
