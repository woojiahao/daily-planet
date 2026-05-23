package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/woojiahao/daily-planet/bot"
	"github.com/woojiahao/daily-planet/cron"
	"github.com/woojiahao/daily-planet/db"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// if .env doesn't exist, rely on the environment so don't need to panic
	err := godotenv.Load()
	if err != nil {
		fmt.Printf(".env not found, relying on environments\n")
	}

	database, err := db.New()
	if err != nil {
		panic(err)
	}

	bot, err := bot.NewBot(database)
	if err != nil {
		panic(err)
	}

	engine := cron.NewCronEngine(database, bot)
	bot.SetScheduler(engine)

	if err = engine.Start(); err != nil {
		panic(err)
	}

	if err = bot.Run(); err != nil {
		panic(err)
	}

	<-ctx.Done()
	bot.Stop()
	engine.Stop()
}
