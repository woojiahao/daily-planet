package common

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

func DBRoot() string {
	if p := os.Getenv("DB_ROOT"); p != "" {
		return p
	}

	// used for local development
	return "./"
}
