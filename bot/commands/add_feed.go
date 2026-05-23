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
	"github.com/woojiahao/daily-planet/source"
)

var AddFeed = Command{
	Name:        "add",
	Group:       "feed",
	Description: "Add a feed to the Daily Planet",
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

		err := context.Database.WithTransaction(func(tx db.Database) error {
			// Load the initial feed and cache results so that initial print won't be all spam
			feed, err := source.LoadFeed(url)
			if err != nil {
				return err
			}

			dbFeeds, err := tx.Feed.Insert(models.FeedInsert{
				ConfigurationID: configurationID,
				URL:             url,
				FeedType:        string(feed.EngineType),
			})
			if err != nil {
				return err
			}
			dbFeed := dbFeeds[0]
			cacheKey := models.NewCacheKey(configurationID, dbFeed.ID)

			// ensure that cache isn't duplicated
			// TODO(woojiahao): this would benefit from a proper UNIQUE key on (configuration_id, feed_id) but this suffices for now
			existingCache, err := tx.Cache.All(cacheKey)
			if err != nil {
				return err
			}

			_, newArticleKeys := source.FetchNewArticles(feed, existingCache)

			var cacheInserts []models.CacheInsert
			// for all articles, bulk insert them into the cache
			for _, articleKey := range newArticleKeys {
				cacheInserts = append(cacheInserts, models.CacheInsert{
					CacheKey:   cacheKey,
					ArticleKey: articleKey,
				})
			}
			err = tx.Cache.Insert(cacheInserts...)
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
					"Feed added",
					fmt.Sprintf("Added feed %s to source", url),
					common.ColorGreen,
				),
				apperrors.ErrLoadFeedFailed: helpers.CreateSimpleEmbed(
					"Feed could not be loaded",
					fmt.Sprintf("Failed to load feed %s into source.\nVerify that the feed is well-formed.", url),
					common.ColorRed,
				),
				apperrors.ErrFeedDBError: helpers.CreateSimpleEmbed(
					"Feed NOT added",
					fmt.Sprintf("Failed to add feed %s to source. Try again.", url),
					common.ColorRed,
				),
				apperrors.ErrCacheDBError: helpers.CreateSimpleEmbed(
					"Feed NOT added",
					"Failed to save feed articles into cache",
					common.ColorRed,
				),
			},
		)
	},
}
