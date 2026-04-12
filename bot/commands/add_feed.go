package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-sqlite3"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/source"
)

var AddFeed = Command{
	Name:        "add-feed",
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
		url := strings.Trim(context.Interaction.ApplicationCommandData().Options[0].StringValue(), " ")

		// Load the initial feed and cache results so that initial print won't be all spam
		feed, err := source.LoadFeed(url)
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Feed could not be loaded",
				fmt.Sprintf("Failed to load feed %s into source.\nVerify that the feed is well-formed.", url, url),
				helpers.ColorRed,
			)
		}

		dbFeed, err := context.Database.Feed.InsertOne(context.CallerConfiguration.ID, url, string(feed.EngineType))
		if err != nil {
			fmt.Printf("err is %v\n", err)
			if sqlite3Err, ok := err.(sqlite3.Error); ok {
				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
					return helpers.CreateSimpleEmbed(
						"Feed NOT added",
						fmt.Sprintf("Source %s already exists.\n\nUse `/list-feeds` to locate it or `/enable-feed %s` to enable it if it has been disabled.", url),
						helpers.ColorRed,
					)
				}
			}
			return helpers.CreateSimpleEmbed(
				"Feed NOT added",
				fmt.Sprintf("Failed to add feed %s to source. Try again.", url),
				helpers.ColorRed,
			)
		}

		var articleKeys []string
		// for all articles, bulk insert them into the cache
		for _, article := range feed.Articles {
			articleKeys = append(articleKeys, string(article.GetKey()))
		}
		err = context.Database.Cache.InsertManyWithSameConfigurationIDAndFeedID(context.CallerConfiguration.ID, dbFeed.ID, articleKeys)
		if err != nil {
			fmt.Printf("err is %v\n", err)
			return helpers.CreateSimpleEmbed(
				"Feed NOT added",
				"Failed to load feed articles into cache",
				helpers.ColorRed,
			)
		}

		return helpers.CreateSimpleEmbed(
			"Feed added",
			fmt.Sprintf("Added feed %s to source", url),
			helpers.ColorGreen,
		)
	},
}
