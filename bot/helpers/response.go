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

type Embed struct {
	Title       string
	Description string
	Color       Color
	Fields      []*discordgo.MessageEmbedField
	Footer      *discordgo.MessageEmbedFooter
}

func (e Embed) toMessageEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       e.Title,
		Description: e.Description,
		Color:       int(e.Color),
		Fields:      e.Fields,
		Footer:      e.Footer,
	}
}

func CreateEmbed(embed Embed) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed.toMessageEmbed(),
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

func CreateEphemeralEmbed(embed Embed) *discordgo.InteractionResponse {
	return createEphemeralResponse(CreateEmbed(embed))
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

func SendMessage(session *discordgo.Session, interaction *discordgo.InteractionCreate, content string) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func SendEmbed(session *discordgo.Session, interaction *discordgo.InteractionCreate, embed Embed) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed.toMessageEmbed(),
			},
		},
	})
}
