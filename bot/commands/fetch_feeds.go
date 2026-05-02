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

func fetchFeedsAlgorithm(context context.CommandContext) {
	// Fetching algorithm:
	// 1. Fetch all enabled feeds for this configuration
	// 2. Map the feed ID -> feed / feed URL from the enabled feed
	// 3. Retrieve all of the caches for the enabled feeds
	// 4. Fetch all feeds from all sources concurrently
	// 5. For each feed retrieved,
	// 	a. For each article retrieved,
	// 		1) Check if article exists within cache
	// 		2) If article exists within cache, skip over it
	// 		3) If article does not exist within cache, add to be inserted and add to new articles

	enabledFeeds, err := context.Database.Feed.AllEnabledByConfigurationID(context.CallerConfiguration.ID)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		helpers.SendFollowupSimpleEmbed(
			context.Session,
			context.Interaction,
			"Failed to fetch feeds",
			"Failed to fetch feeds for this source.",
			helpers.ColorRed,
		)
		return
	}

	fmt.Println("enabled feeds:")
	for _, feed := range enabledFeeds {
		fmt.Printf("- %s\n", feed.URL)
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
		fmt.Printf("err is %v\n", err)
		helpers.SendFollowupSimpleEmbed(
			context.Session,
			context.Interaction,
			"Failed to fetch feeds",
			"Failed to fetch cache for feeds in this source.",
			helpers.ColorRed,
		)
		return
	}

	feedCaches := make(map[models.FeedID][]models.Cache)
	for _, cache := range caches {
		feedCaches[cache.FeedID] = append(feedCaches[cache.FeedID], cache)
	}

	fmt.Println("feed caches:")
	for k, v := range feedCaches {
		fmt.Printf("cache for %d with %d entries\n", k, len(v))
	}

	urls := slices.Sorted(maps.Values(feedURLs))
	feeds := source.BulkLoadFeeds(urls)

	feedsByURLs := make(map[string]source.Feed)
	for i, url := range urls {
		feedsByURLs[url] = feeds[i]
	}

	var insertCacheKeys []models.CacheKey
	var insertArticleKeys []string
	var allArticles []source.Article
	for feedID := range feedURLs {
		feedCache := feedCaches[feedID]

		newArticles, newArticleKeys := helpers.FetchNewArticles(
			feedsByURLs[feedURLs[feedID]],
			feedCache,
		)

		allArticles = append(allArticles, newArticles...)
		for _, articleKey := range newArticleKeys {
			cacheKey := models.NewCacheKey(context.CallerConfiguration.ID, feedID)
			insertCacheKeys = append(insertCacheKeys, cacheKey)
			insertArticleKeys = append(insertArticleKeys, articleKey)
		}
	}

	if len(allArticles) != 0 {
		err = context.Database.Cache.InsertMany(insertCacheKeys, insertArticleKeys)
		if err != nil {
			fmt.Printf("%v\n", err)
			helpers.SendFollowupSimpleEmbed(
				context.Session,
				context.Interaction,
				"Failed to save cache",
				"Failed to save articles into cache in this source",
				helpers.ColorRed,
			)
			return
		}

		var articleStrings []string
		for _, article := range allArticles {
			articleStrings = append(articleStrings, fmt.Sprintf("- [%s](%s)", article.Title, article.Link))
		}

		fmt.Printf("articleStrings total length %d\n", len(articleStrings))

		var groupedArticleStrings [][]string
		groupedArticleStrings = append(groupedArticleStrings, []string{})
		acc := 0
		const limit = 3500
		for _, articleString := range articleStrings {
			// include \n at the end as well
			if acc+len(articleString)+1 > limit {
				// split out
				groupedArticleStrings = append(groupedArticleStrings, []string{})
				acc = len(articleString) + 1
			} else {
				acc += len(articleString) + 1
			}
			groupedArticleStrings[len(groupedArticleStrings)-1] = append(groupedArticleStrings[len(groupedArticleStrings)-1], articleString)
		}

		fmt.Printf("gropued article strings length %d\n", len(groupedArticleStrings))

		if len(groupedArticleStrings) > 0 {
			for _, group := range groupedArticleStrings {
				helpers.SendFollowupSimpleEmbed(
					context.Session,
					context.Interaction,
					"Feeds fetched",
					strings.Join(group, "\n"),
					helpers.ColorBlue,
				)
			}
		}

		return
	}

	helpers.SendFollowupSimpleEmbed(
		context.Session,
		context.Interaction,
		"Feeds fetched",
		"All feeds fetched but no new articles found",
		helpers.ColorBlue,
	)
}

var FetchFeeds = Command{
	Name:        "fetch-all",
	Group:       "feed",
	Description: "Updates every feed",
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		// Given that this algorithm takes a while to complete if everything is uncached, we will defer the response
		go fetchFeedsAlgorithm(context)
		return helpers.CreateDeferredResponse()
	},
}
