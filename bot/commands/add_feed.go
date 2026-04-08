package commands

import (
	"fmt"

	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/source"
)

func AddFeed(context CommandContext) {
	// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
	url := context.Interaction.ApplicationCommandData().Options[0].StringValue()

	// Load the initial feed and cache results so that initial print won't be all spam
	feed, err := source.LoadFeed(url)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Feed could not be loaded",
			fmt.Sprintf("Failed to load feed %s into source.\nVerify that the feed is well-formed.", url),
			helpers.ColorRed,
		)
		return
	}

	dbFeed, err := context.Database.Feed.InsertOne(context.CallerConfiguration.SnowflakeID, url, string(feed.EngineType))
	if err != nil {
		fmt.Printf("err is %v\n", err)
		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Feed NOT added",
			fmt.Sprintf("Failed to add feed %s to source. Try again.", url),
			helpers.ColorRed,
		)
		return
	}

	var configurationIDs []int
	var feedIDs []int
	var articleKeys []string
	// for all articles, bulk insert them into the cache
	for _, article := range feed.Articles {
		configurationIDs = append(configurationIDs, context.CallerConfiguration.ID)
		feedIDs = append(feedIDs, dbFeed.ID)
		articleKeys = append(articleKeys, string(article.GetKey()))
	}
	err = context.Database.Cache.InsertMany(configurationIDs, feedIDs, articleKeys)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		helpers.SendEmbed(
			context.Session,
			context.Interaction,
			"Feed NOT added",
			"Failed to load feed articles into cache",
			helpers.ColorRed,
		)
		return
	}

	helpers.SendEmbed(
		context.Session,
		context.Interaction,
		"Feed added",
		fmt.Sprintf("Added feed %s to source", url),
		helpers.ColorGreen,
	)
}
