package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/goccy/go-yaml"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db/models"
	"github.com/woojiahao/daily-planet/source"
)

var UploadFeeds = Command{
	Name:        "upload",
	Group:       "feed",
	Description: "Bulk upload feeds into this source via a .yaml file with a 'feeds' key",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Name:        "feeds_file",
			Description: "Feeds .yaml file",
			Required:    true,
		},
	},
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
		yamlConfig := helpers.GetRequiredOption[[]byte](context, "feeds_file")

		var rawFeedData struct {
			Feeds []string `yaml:"feeds"`
		}

		if err := yaml.Unmarshal(yamlConfig, &rawFeedData); err != nil {
			panic(err)
		}

		urls := rawFeedData.Feeds

		feeds := source.BulkLoadFeeds(urls)

		var feedURLs []string
		var feedTypes []string
		for _, feed := range feeds {
			// if no title, means it didn't load so skip
			if feed.Title == "" {
				continue
			}
			feedURLs = append(feedURLs, feed.RawLink)
			feedTypes = append(feedTypes, string(feed.EngineType))
		}
		fmt.Printf("%v\n", feedTypes)

		dbFeeds, err := context.Database.Feed.InsertManyWithSameConfigurationID(context.CallerConfiguration.ID, feedURLs, feedTypes)
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Feeds NOT uploaded",
				"Failed to upload feeds to source. Try again.",
				common.ColorRed,
			)
		}

		dbFeedIDsByURL := make(map[string]models.FeedID)
		for _, f := range dbFeeds {
			dbFeedIDsByURL[f.URL] = f.ID
		}

		var articleKeys []string
		var cacheKeys []models.CacheKey
		// for all articles, bulk insert them into the cache
		for _, feed := range feeds {
			for _, article := range feed.Articles {
				articleKeys = append(articleKeys, string(article.GetKey()))
				cacheKeys = append(cacheKeys, models.NewCacheKey(context.CallerConfiguration.ID, dbFeedIDsByURL[feed.RawLink]))
			}
		}
		err = context.Database.Cache.InsertMany(
			cacheKeys,
			articleKeys,
		)
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Feeds NOT uploaded",
				"Failed to add feed articles into cache",
				common.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feed added",
			"Uploaded feeds to source",
			common.ColorGreen,
		)
	},
}
