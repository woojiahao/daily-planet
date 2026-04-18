package helpers

import "os"

func BotToken() string {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		panic("missing DISCORD_TOKEN in environment")
	}
	return discordToken
}

func TestGuildID() string {
	return os.Getenv("TEST_GUILD_ID")
}
