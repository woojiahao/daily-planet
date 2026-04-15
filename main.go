package main

import (
	"github.com/joho/godotenv"
	"github.com/woojiahao/daily-planet/bot"
	"github.com/woojiahao/daily-planet/cron"
	"github.com/woojiahao/daily-planet/db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	database, err := db.New()
	if err != nil {
		panic(err)
	}

	cron.NewCronEngine(database)

	bot := bot.NewBot(database)
	bot.Run()
}
