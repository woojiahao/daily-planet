// Package bot contains the implementation for the Discord bot aspects of daily planet.
package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/woojiahao/daily-planet/bot/commands"
)

func botToken() string {
	discord_token := os.Getenv("DISCORD_TOKEN")
	if discord_token == "" {
		panic("missing DISCORD_TOKEN in environment")
	}
	return discord_token
}

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandName := i.ApplicationCommandData().Name
	if v, ok := commands.CommandHandlerMapping[commandName]; !ok {
		return
	} else {
		v(s, i)
	}
}

func Run() {
	discord, err := discordgo.New("Bot " + botToken())
	if err != nil {
		panic(err)
	}

	discord.AddHandler(interactionHandler)

	err = discord.Open()
	if err != nil {
		panic(err)
	}

	for _, command := range commands.CommandDefinitions {
		_, err = discord.ApplicationCommandCreate(discord.State.User.ID, "", command)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("bot is running\n")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	discord.Close()
}
