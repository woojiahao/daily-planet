package main

import (
	"github.com/joho/godotenv"
	"github.com/woojiahao/daily-planet/bot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	bot.Run()
}
