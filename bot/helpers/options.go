package helpers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
)

func GetOption[T any](context context.CommandContext, name string) (T, bool) {
	// this is a small hack-y approach because a method cannot contain a generic, so we cannot inline this into the CommandContext as a method
	var zero T
	data := context.Interaction.ApplicationCommandData()

	var find func(opts []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption
	find = func(opts []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
		for _, opt := range opts {
			if opt.Name == name {
				return opt
			}

			if len(opt.Options) > 0 {
				if found := find(opt.Options); found != nil {
					return found
				}
			}
		}

		return nil
	}

	opt := find(data.Options)

	switch opt.Type {
	case discordgo.ApplicationCommandOptionString:
		return any(opt.StringValue()).(T), true
	case discordgo.ApplicationCommandOptionInteger:
		return any(opt.IntValue()).(T), true
	case discordgo.ApplicationCommandOptionBoolean:
		return any(opt.BoolValue()).(T), true
	case discordgo.ApplicationCommandOptionNumber:
		return any(opt.FloatValue()).(T), true
	case discordgo.ApplicationCommandOptionUser:
		return any(opt.UserValue(context.Session)).(T), true
	case discordgo.ApplicationCommandOptionChannel:
		return any(opt.ChannelValue(context.Session)).(T), true
	case discordgo.ApplicationCommandOptionRole:
		return any(opt.RoleValue(context.Session, context.Interaction.GuildID)).(T), true
	}

	return zero, false
}

// GetRequiredOption assumes that the option is always provided so no flag is required
func GetRequiredOption[T any](context context.CommandContext, name string) T {
	// this is a small hack-y approach because a method cannot contain a generic, so we cannot inline this into the CommandContext as a method
	var zero T
	data := context.Interaction.ApplicationCommandData()

	var find func(opts []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption
	find = func(opts []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
		for _, opt := range opts {
			if opt.Name == name {
				return opt
			}

			if len(opt.Options) > 0 {
				if found := find(opt.Options); found != nil {
					return found
				}
			}
		}

		return nil
	}

	opt := find(data.Options)

	switch opt.Type {
	case discordgo.ApplicationCommandOptionString:
		return any(opt.StringValue()).(T)
	case discordgo.ApplicationCommandOptionInteger:
		return any(opt.IntValue()).(T)
	case discordgo.ApplicationCommandOptionBoolean:
		return any(opt.BoolValue()).(T)
	case discordgo.ApplicationCommandOptionNumber:
		return any(opt.FloatValue()).(T)
	case discordgo.ApplicationCommandOptionUser:
		return any(opt.UserValue(context.Session)).(T)
	case discordgo.ApplicationCommandOptionChannel:
		return any(opt.ChannelValue(context.Session)).(T)
	case discordgo.ApplicationCommandOptionRole:
		return any(opt.RoleValue(context.Session, context.Interaction.GuildID)).(T)
	}

	return zero
}
