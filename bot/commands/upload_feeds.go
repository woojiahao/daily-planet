package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/goccy/go-yaml"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db"
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
		yamlConfig := helpers.GetRequiredOption[[]byte](context, "feeds_file")
		configurationID := context.CallerConfiguration.ID

		go func() {
			context.Database.WithTransaction(func(tx db.Database) error {
				var rawFeedData struct {
					Feeds []string `yaml:"feeds"`
				}

				if err := yaml.Unmarshal(yamlConfig, &rawFeedData); err != nil {
					helpers.SendFollowupSimpleEmbed(
						context.Session,
						context.Interaction,
						"Failed to load YAML",
						"Invalid YAML provided",
						common.ColorRed,
					)
					return err
				}

				urls := rawFeedData.Feeds

				feeds := source.BulkLoadFeeds(urls)

				var feedInserts []models.FeedInsert
				for _, feed := range feeds {
					// if no title, means it didn't load so skip
					if feed.Title == "" {
						continue
					}
					feedInserts = append(feedInserts, models.FeedInsert{
						ConfigurationID: configurationID,
						URL:             feed.RawLink,
						FeedType:        string(feed.EngineType),
					})
				}

				// if no feeds got inserted, means that nothing was new, so we might have overlapping caches
				// calculate the cache difference and insert
				dbFeeds, err := tx.Feed.Insert(feedInserts...)
				if err != nil {
					fmt.Printf("err is %v\n", err)
					helpers.SendFollowupSimpleEmbed(
						context.Session,
						context.Interaction,
						"Feeds NOT uploaded",
						"Failed to upload feeds to source. Try again.",
						common.ColorRed,
					)
					return err
				}

				dbFeedIDsByURL := make(map[string]models.FeedID)
				for _, f := range dbFeeds {
					dbFeedIDsByURL[f.URL] = f.ID
				}

				var cacheInserts []models.CacheInsert
				// for all articles, bulk insert them into the cache
				for _, feed := range feeds {
					existingCache, err := tx.Cache.All(models.NewCacheKey(configurationID, dbFeedIDsByURL[feed.RawLink]))
					if err != nil {
						fmt.Printf("err is %v\n", err)
						helpers.SendFollowupSimpleEmbed(
							context.Session,
							context.Interaction,
							"Feeds NOT uploaded",
							"Failed to upload feeds to source because cache could not be found. Try again.",
							common.ColorRed,
						)
						return err
					}

					// TODO(woojiahao): this would benefit from a proper UNIQUE key on (configuration_id, feed_id) but this suffices for now
					_, newArticleKeys := source.FetchNewArticles(feed, existingCache)
					for _, articleKey := range newArticleKeys {
						cacheInserts = append(cacheInserts, models.CacheInsert{
							ArticleKey: articleKey,
							CacheKey:   models.NewCacheKey(configurationID, dbFeedIDsByURL[feed.RawLink]),
						})
					}
				}
				err = tx.Cache.Insert(cacheInserts...)
				if err != nil {
					fmt.Printf("err is %v\n", err)
					helpers.SendFollowupSimpleEmbed(
						context.Session,
						context.Interaction,
						"Feeds NOT uploaded",
						"Failed to add feed articles into cache",
						common.ColorRed,
					)
					return err
				}

				helpers.SendFollowupSimpleEmbed(
					context.Session,
					context.Interaction,
					"Feed added",
					"Uploaded feeds to source",
					common.ColorGreen,
				)

				return nil
			})
		}()

		return helpers.CreateDeferredResponse()
	},
}
