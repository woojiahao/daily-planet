package commands

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/ds"
	"github.com/woojiahao/daily-planet/source"
)

var FetchFeed = Command{
	Name:        "fetch-feed",
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
		url := helpers.GetRequiredOption[string](context, "url")

		feed, err := source.LoadFeed(url)
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Feed could not be loaded",
				fmt.Sprintf("Failed to load feed %s into source.\nVerify that the feed is well-formed.", url, url),
				helpers.ColorRed,
			)
		}

		dbFeed, err := context.Database.Feed.OneByConfigurationIDAndURL(context.CallerConfiguration.ID, url)
		if err != nil {
			if err == sql.ErrNoRows {
				return helpers.CreateSimpleEmbed(
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse `/list-feeds` to verify that it exists in this source.", url),
					helpers.ColorRed,
				)
			}
			return helpers.CreateSimpleEmbed(
				"Failed to fetch feed",
				fmt.Sprintf("Failed to fetch feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		cachedArticles, err := context.Database.Cache.AllByConfigurationIDAndFeedID(context.CallerConfiguration.ID, dbFeed.ID)
		if err != nil {
			return helpers.CreateSimpleEmbed(
				"Failed to fetch feed cache",
				fmt.Sprintf("Failed to fetch cache for feed %s", url),
				helpers.ColorRed,
			)
		}

		cachedArticleKeys := ds.NewSet[source.ArticleKey]()
		for _, article := range cachedArticles {
			cachedArticleKeys.Add(source.ArticleKey(article.ArticleKey))
		}

		var newArticles []source.Article
		var newArticleKeys []string
		for _, article := range feed.Articles {
			if !cachedArticleKeys.Contains(article.GetKey()) {
				// new article to cache and print
				newArticles = append(newArticles, article)
				newArticleKeys = append(newArticleKeys, string(article.GetKey()))
			}
		}

		err = context.Database.Cache.InsertManyWithSameConfigurationIDAndFeedID(
			context.CallerConfiguration.ID,
			dbFeed.ID,
			newArticleKeys,
		)
		if err != nil {
			return helpers.CreateSimpleEmbed(
				"Failed to save cache",
				fmt.Sprintf("Failed to save articles into cache for feed %s", url),
				helpers.ColorRed,
			)
		}

		var articleStrings []string
		for _, article := range newArticles {
			articleStrings = append(articleStrings, fmt.Sprintf("- [%s](%s)", article.Title, article.Link))
		}

		return helpers.CreateSimpleEmbed(
			"Feed fetched",
			strings.Join(articleStrings, "\n"),
			helpers.ColorBlue,
		)
	},
}
