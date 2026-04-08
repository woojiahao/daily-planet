package commands

import "github.com/bwmarrin/discordgo"

var DeleteFeed = Command{
	Name:        "delete-feed",
	Description: "Deletes a feed from the Daily Planet",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "Feed URL to delete",
			Required:    true,
		},
	},
	Handler: func(context CommandContext) {},
}
