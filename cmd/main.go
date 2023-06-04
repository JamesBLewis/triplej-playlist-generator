package main

import (
	"context"
	"fmt"
	"log"

	"github.com/JamesBLewis/triplej-playlist-generator/cmd/config"
	"github.com/JamesBLewis/triplej-playlist-generator/internal"
)

func main() {
	fmt.Println("🤖Triplej RunBot is running...")
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", err)
	}
	bot := internal.NewBot(cfg)
	err = bot.Run(ctx)
	if err != nil {
		log.Fatal("bot ran into an error while running", err)
	}
	fmt.Println("🤖Done.")
}
