package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
	"github.com/woojiahao/daily-planet/bot/helpers"
)

var SetChannel = Command{
	Name:        "set-channel",
	Description: "Set the configuration channel for servers to the channel this command is invoked in",
	Group:       CommandGroupConfiguration,
	Handler: func(context context.CommandContext) *discordgo.InteractionResponse {
		if context.Interaction.GuildID == "" {
			return helpers.CreateSimpleEmbed("Channel not required", "You do not need to set the channel in a DM. We will send updates directly to your DM.", helpers.ColorBlue)
		}

		currentChannelID := context.Interaction.ChannelID
		context.Database.Configuration.UpdateOneByID(context.CallerConfiguration.ID, nil, &currentChannelID, nil, nil)
		return helpers.CreateSimpleEmbed("Channel set", fmt.Sprintf("Channel for this source set to <#%s>", currentChannelID), helpers.ColorGreen)
	},
}
