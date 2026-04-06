// Package helpers provides helpers over the discordgo API to reduce duplication
package helpers

import "github.com/bwmarrin/discordgo"

func SendMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func SendEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string, color int) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       color,
				},
			},
		},
	})
}

func foo() {
}
