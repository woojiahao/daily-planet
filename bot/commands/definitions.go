package commands

import "github.com/bwmarrin/discordgo"

type CommandName string

const (
	CommandNamePing       CommandName = "ping"
	CommandNameListFeeds  CommandName = "list-feeds"
	CommandNameAddFeed    CommandName = "add-feed"
	CommandNameDeleteFeed CommandName = "delete-feed"
)

var CommandHandlerMapping = map[CommandName]CommandHandler{
	CommandNamePing:      Ping,
	CommandNameListFeeds: ListFeeds,
	CommandNameAddFeed:   AddFeed,
}

var CommandDefinitions = []*discordgo.ApplicationCommand{
	{
		Name:        string(CommandNamePing),
		Description: "Check if the bot is alive",
	},
	{
		Name:        string(CommandNameListFeeds),
		Description: "List current feed for the Daily Planet",
	},
	// TODO(woojiahao): allow cron to be configured here as an optional parameter
	{
		Name:        string(CommandNameAddFeed),
		Description: "Add a feed to the Daily Planet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "Feed URL to add",
				Required:    true,
			},
		},
	},
	{
		Name:        string(CommandNameDeleteFeed),
		Description: "Deletes a feed from the Daily Planet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "Feed URL to delete",
				Required:    true,
			},
		},
	},
}
