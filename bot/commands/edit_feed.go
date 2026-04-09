package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var EditFeed = Command{
	Name:        "edit-feed",
	Description: "Edit a feed in the source",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "Feed URL to edit",
			Required:    true,
		},
	},
	Handler: func(context CommandContext) *discordgo.InteractionResponse {
		// TODO(woojiahao): wrap these in a transaction instead of separating the API calls
		url := strings.Trim(context.Interaction.ApplicationCommandData().Options[0].StringValue(), " ")
		feed, err := context.Database.Feed.OneByConfigurationIDAndURL(context.CallerConfiguration.ID, url)
		if err != nil {
			if err == sql.ErrNoRows {
				return helpers.CreateEmbed(
					"Feed not found",
					fmt.Sprintf("Failed to fetch feed by URL %s as it does not exist.\n\nUse /list-feeds to verify that it exists in this source.", url),
					helpers.ColorRed,
				)
			}
			return helpers.CreateEmbed(
				"Failed to fetch feed",
				fmt.Sprintf("Failed to fetch feed by URL %s. Try again", url),
				helpers.ColorRed,
			)
		}

		cronSchedule := context.CallerConfiguration.CronSchedule
		if feed.CronSchedule.Valid {
			cronSchedule = feed.CronSchedule.String
		}

		return helpers.CreateModal(
			"edit_feed_"+context.Interaction.Member.User.ID,
			"Edit feed",
			[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "feed_url",
							Label:    "Feed URL",
							Style:    discordgo.TextInputShort,
							Value:    feed.URL,
							Required: true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "cron_schedule",
							Label:       "Cron schedule",
							Style:       discordgo.TextInputShort,
							Placeholder: cronSchedule,
							Required:    false,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "disabled",
							Label:    "Disabled (true/false)",
							Style:    discordgo.TextInputShort,
							Value:    strconv.FormatBool(feed.Disabled),
						},
					},
				},
			})
	},
}
