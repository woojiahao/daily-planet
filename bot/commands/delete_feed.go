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
	"github.com/woojiahao/daily-planet/db/models"
)

var DeleteFeed = Command{
	Name:        "delete",
	Group:       "feed",
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
		url := strings.Trim(helpers.GetRequiredOption[string](context, "url"), " ")
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			feed, err := tx.Feed.OneByKey(models.NewFeedKey(configurationID, url))
			if err != nil {
				return err
			}

			err = tx.Feed.DeleteOneByID(feed.ID)
			if err != nil {
				return err
			}

			return nil
		})

		return common.SwitchErrorWithDefaultFunc(
			err,
			helpers.UnknownErrorHandler(),
			map[error]*discordgo.InteractionResponse{
				nil: helpers.CreateSimpleEmbed(
					"Feed deleted",
					fmt.Sprintf("Feed %s has been deleted from this source", url),
					common.ColorGreen,
				),
				apperrors.ErrFeedNotFound: helpers.CreateSimpleEmbed(
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse `/list-feeds` to verify that it exists in this source.", url),
					common.ColorRed,
				),
				apperrors.ErrFeedDBError: helpers.CreateSimpleEmbed(
					"Failed to delete feed",
					fmt.Sprintf("Failed to delete feed by URL %s. Try again", url),
					common.ColorRed,
				),
			},
		)
	},
}
