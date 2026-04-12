package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var ListFeeds = Command{
	Name:        "list-feeds",
	Description: "List current feed for the Daily Planet",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		feeds, err := context.Database.Feed.All()
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Failed to load feeds",
				"The Daily Planet failed to load feeds for this source",
				helpers.ColorRed,
			)
		}

		var enabledFeeds []string
		var disabledFeeds []string
		for _, feed := range feeds {
			if feed.Disabled {
				disabledFeeds = append(disabledFeeds, "- "+feed.URL)
			} else {
				enabledFeeds = append(enabledFeeds, "- "+feed.URL)
			}
		}

		var fields []*discordgo.MessageEmbedField
		if len(enabledFeeds) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Enabled",
				Value: strings.Join(enabledFeeds, "\n"),
			})
		}

		if len(disabledFeeds) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Disabled",
				Value: strings.Join(disabledFeeds, "\n"),
			})
		}

		return helpers.CreateEmbed(
			helpers.Embed{
				Title:  "Feeds fetched",
				Color:  helpers.ColorBlue,
				Fields: fields,
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf(
						"Feeds: %d; enabled: %d; disabled: %d",
						len(feeds),
						len(enabledFeeds),
						len(disabledFeeds),
					),
				},
			},
		)
	},
}
