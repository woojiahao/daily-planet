package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/source"
)

var FetchFeeds = Command{
	Name:        "fetch-all",
	Group:       "feed",
	Description: "Updates every feed",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		// Given that this algorithm takes a while to complete if everything is uncached, we will defer the response
		go func() {
			source.FetchFeedsAlgorithmWrapper(
				context.CallerConfiguration.ID,
				context.Database,
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
		}()
		return helpers.CreateDeferredResponse()
	},
}
