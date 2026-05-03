package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-sqlite3"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/source"
)

var (
	errAddFeedLoadFeedFailed = errors.New("failed to load feed")
	errAddFeedFeedExists     = errors.New("feed already added")
	errAddFeedDBIssue        = errors.New("database error occurred")
	errAddFeedCacheNotAdded  = errors.New("cache not added")
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
		// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
		url := strings.Trim(helpers.GetRequiredOption[string](context, "url"), " ")
		configurationID := context.CallerConfiguration.ID

		err := context.Database.WithTransaction(func(tx db.Database) error {
			// Load the initial feed and cache results so that initial print won't be all spam
			feed, err := source.LoadFeed(url)
			if err != nil {
				return errAddFeedLoadFeedFailed
			}

			dbFeed, err := tx.Feed.InsertOne(configurationID, url, string(feed.EngineType))
			if err != nil {
				if sqlite3Err, ok := err.(sqlite3.Error); ok {
					if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
						return errAddFeedFeedExists
					}
				}
				return errAddFeedDBIssue
			}

			var articleKeys []string
			// for all articles, bulk insert them into the cache
			for _, article := range feed.Articles {
				articleKeys = append(articleKeys, string(article.GetKey()))
			}
			err = tx.Cache.InsertManyWithSameKey(
				models.NewCacheKey(context.CallerConfiguration.ID, dbFeed.ID),
				articleKeys,
			)
			if err != nil {
				return errAddFeedCacheNotAdded
			}

			return nil
		})

		return common.SwitchError(err, map[error]*discordgo.InteractionResponse{
			nil: helpers.CreateSimpleEmbed(
				"Feed added",
				fmt.Sprintf("Added feed %s to source", url),
				common.ColorGreen,
			),
			errAddFeedLoadFeedFailed: helpers.CreateSimpleEmbed(
				"Feed could not be loaded",
				fmt.Sprintf("Failed to load feed %s into source.\nVerify that the feed is well-formed.", url),
				common.ColorRed,
			),
			errAddFeedFeedExists: helpers.CreateSimpleEmbed(
				"Feed NOT added",
				fmt.Sprintf("Source %s already exists.\n\nUse `/list-feeds` to locate it or `/enable-feed %s` to enable it if it has been disabled.", url, url),
				common.ColorRed,
			),
			errAddFeedDBIssue: helpers.CreateSimpleEmbed(
				"Feed NOT added",
				fmt.Sprintf("Failed to add feed %s to source. Try again.", url),
				common.ColorRed,
			),
			errAddFeedCacheNotAdded: helpers.CreateSimpleEmbed(
				"Feed NOT added",
				"Failed to load feed articles into cache",
				common.ColorRed,
			),
		})
	},
}
