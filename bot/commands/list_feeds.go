package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
)

var ListFeeds = Command{
	Name:        "list",
	Group:       "feed",
	Description: "List current feed for the Daily Planet",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			feeds, err := tx.Feed.AllByConfigurationID(configurationID)
			if err != nil {
				return err
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

			// TODO(woojiahao): This is super jank, but we basically cannot pass the values from within this scope out so
			// we directly send the message here
			helpers.SendEmbed(
				context.Session,
				context.Interaction,
				helpers.Embed{
					Title:  "Feeds fetched",
					Color:  common.ColorBlue,
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

			return nil
		})

		// skip the default handling because nil is not properly handled
		return common.SwitchError(
			err,
			map[error]*discordgo.InteractionResponse{
				apperrors.ErrFeedDBError: helpers.CreateEphemeralSimpleEmbed(
					"Failed to fetch feeds",
					"Failed to fetch feeds in source.",
					common.ColorRed,
				),
			})
	},
}
