package helpers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/context"
)

func getNestedOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range opts {
		if opt.Name == name {
			return opt
		}

		if len(opt.Options) > 0 {
			if found := getNestedOption(opt.Options, name); found != nil {
				return found
			}
		}
	}

	return nil
}

func GetOption[T any](context context.CommandContext, name string) (T, bool) {
	// this is a small hack-y approach because a method cannot contain a generic, so we cannot inline this into the CommandContext as a method
	var zero T
	data := context.Interaction.ApplicationCommandData()

	opt := getNestedOption(data.Options, name)

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
	case discordgo.ApplicationCommandOptionAttachment:
		attachmentID := opt.Value.(string)
		attachmentURL := data.Resolved.Attachments[attachmentID].URL
		res, err := http.DefaultClient.Get(attachmentURL)
		if err != nil {
			fmt.Println(errors.New("could not get attachment"))
			return zero, false
		}
		if res.StatusCode != 200 {
			fmt.Println(errors.New("could not get attachment"))
			return zero, false
		}
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(errors.New("could not get attachment"))
			return zero, false
		}
		return any(data).(T), true
	}

	return zero, false
}

// GetRequiredOption assumes that the option is always provided so no flag is required
func GetRequiredOption[T any](context context.CommandContext, name string) T {
	value, _ := GetOption[T](context, name)
	return value
}
