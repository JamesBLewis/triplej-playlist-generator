package main

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/JamesBLewis/triplej-playlist-generator/cmd/config"
	"github.com/JamesBLewis/triplej-playlist-generator/internal"
)

func main() {
	fmt.Println("🤖Triplej RunBot is running...")
	ctx := context.Background()
	logger, err := config.BuildLogger()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Fatalf("failed to sync logs: %v", err)
		}
	}(logger)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.NamedError("ConfigError", err))
	}
	bot := internal.NewBot(cfg, logger)
	err = bot.Run(ctx)
	if err != nil {
		logger.Fatal("bot ran into an error while running", zap.NamedError("RuntimeError", err))
	}
	fmt.Println("🤖Done.")
}
