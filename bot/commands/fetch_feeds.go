package commands

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/source"
)

var FetchFeeds = Command{
	Name:        "fetch-all",
	Group:       "feed",
	Description: "Updates every feed",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		enabledFeeds, err := context.Database.Feed.AllEnabled()
		if err != nil {
			return helpers.CreateSimpleEmbed("Failed to fetch feeds", "Failed to fetch feeds for this source.", helpers.ColorRed)
		}

		feedURLs := make(map[models.FeedID]string)
		feedMap := make(map[models.FeedID]models.Feed)
		var cacheKeys []models.CacheKey
		for _, feed := range enabledFeeds {
			feedMap[feed.ID] = feed
			feedURLs[feed.ID] = feed.URL
			cacheKeys = append(cacheKeys, models.NewCacheKey(feed.ConfigurationID, feed.ID))
		}
		caches, err := context.Database.Cache.AllByKeys(cacheKeys)
		if err != nil {
			return helpers.CreateSimpleEmbed("Failed to fetch feeds", "Failed to fetch cache for feeds in this source.", helpers.ColorRed)
		}

		feedCaches := make(map[models.FeedID][]models.Cache)
		for _, cache := range caches {
			feedCaches[cache.FeedID] = append(feedCaches[cache.FeedID], cache)
		}

		feeds := source.BulkLoadFeeds(slices.Collect(maps.Values(feedURLs)))
		feedsByURLs := make(map[string]source.Feed)
		for _, feed := range feeds {
			feedsByURLs[feed.Link] = feed
		}

		var insertCacheKeys []models.CacheKey
		var insertArticleKeys []string
		var allArticles []source.Article
		for feedID, feedCache := range feedCaches {
			newArticles, newArticleKeys := helpers.FetchNewArticles(
				feedsByURLs[feedURLs[feedID]],
				feedCache,
			)
			allArticles = append(allArticles, newArticles...)
			for _, articleKey := range newArticleKeys {
				insertCacheKeys = append(insertCacheKeys, models.NewCacheKey(context.CallerConfiguration.ID, feedID))
				insertArticleKeys = append(insertArticleKeys, articleKey)
			}
		}

		if len(allArticles) != 0 {
			err = context.Database.Cache.InsertMany(insertCacheKeys, insertArticleKeys)
			if err != nil {
				fmt.Printf("%v\n", err)
				return helpers.CreateSimpleEmbed(
					"Failed to save cache",
					"Failed to save articles into cache in this source",
					helpers.ColorRed,
				)
			}

			var articleStrings []string
			for _, article := range allArticles {
				articleStrings = append(articleStrings, fmt.Sprintf("- [%s](%s)", article.Title, article.Link))
			}

			return helpers.CreateSimpleEmbed(
				"Feeds fetched",
				strings.Join(articleStrings, "\n"),
				helpers.ColorBlue,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feeds fetched",
			"All feeds fetched but no new articles found",
			helpers.ColorBlue,
		)
	},
}
