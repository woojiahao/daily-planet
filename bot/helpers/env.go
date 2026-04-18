package helpers

import "os"

func botToken() string {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		panic("missing DISCORD_TOKEN in environment")
	}
	return discordToken
}

func testGuildID() string {
	return os.Getenv("TEST_GUILD_ID")
}
