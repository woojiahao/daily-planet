// Package helpers provides helpers over the discordgo API to reduce duplication
package helpers

import (
	"github.com/bwmarrin/discordgo"
)

func CreateMessage(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}
}

func CreateEmbed(title, description string, color Color) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       int(color),
				},
			},
		},
	}
}

func CreateModal(customID, title string, components []discordgo.MessageComponent) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   customID,
			Title:      title,
			Components: components,
		},
	}
}

func CreateMessageComponent(content string, components []discordgo.MessageComponent) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: components,
		},
	}
}

func CreateEphemeralMessage(content string) *discordgo.InteractionResponse {
	return createEphemeralResponse(CreateMessage(content))
}

func CreateEphemeralEmbed(title, description string, color Color) *discordgo.InteractionResponse {
	return createEphemeralResponse(CreateEmbed(title, description, color))
}

func CreateEphemeralModal(customID, title string, components []discordgo.MessageComponent) *discordgo.InteractionResponse {
	return createEphemeralResponse(CreateModal(customID, title, components))
}

func CreateEphemeralMessageComponent(content string, components []discordgo.MessageComponent) *discordgo.InteractionResponse {
	return createEphemeralResponse(CreateMessageComponent(content, components))
}

func createEphemeralResponse(response *discordgo.InteractionResponse) *discordgo.InteractionResponse {
	response.Data.Flags = discordgo.MessageFlagsEphemeral
	return response
}

func SendMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func SendEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string, color Color) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       int(color),
				},
			},
		},
	})
}
