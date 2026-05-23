package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/source"
)

var FetchFeed = Command{
	Name:        "fetch",
	Group:       "feed",
	Description: "Retrieves the latest articles of a given feed",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "Feed URL to add",
			Required:    true,
		},
	},
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		url := strings.Trim(helpers.GetRequiredOption[string](context, "url"), " ")
		configurationID := context.CallerConfiguration.ID

		go func() {
			context.Database.WithTransaction(func(tx db.Database) error {
				source.FetchFeedAlgorithmWrapper(
					models.NewFeedKey(configurationID, url),
					&tx,
					true,
					func(title, description string, color common.Color) {
						helpers.SendFollowupSimpleEmbed(
							context.Session,
							context.Interaction,
							title,
							description,
							color,
						)
					},
				)

				return nil
			})
		}()

		return helpers.CreateDeferredResponse()
	},
}
