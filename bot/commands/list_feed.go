package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/helpers"
	"github.com/woojiahao/daily-planet/db"
)

func ListFeed(s *discordgo.Session, i *discordgo.InteractionCreate, database *db.Database) {
	feeds, err := database.Feed.All()
	if err != nil {
		fmt.Printf("underlying err is %v\n", err)
		helpers.SendEmbed(
			s,
			i,
			"Failed to load feeds",
			"The Daily Planet failed to load feeds for this source.",
			0x5865F2,
		)
		return
	}

	helpers.SendEmbed(
		s,
		i,
		"Feeds fetched",
		fmt.Sprintf("These are the feeds for the current source: %v", feeds),
		0x5865F2,
	)
}
