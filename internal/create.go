package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/evanj/gcplogs/gcpzap"
	"go.uber.org/zap"

	"github.com/JamesBLewis/triplej-playlist-generator/internal/config"
)

func CreateBot() {
	logger, err := gcpzap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	logger.Info("🤖Triplej RunBot is running...")
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.NamedError("ConfigError", err))
	}
	bot := NewBot(cfg, logger)
	err = bot.Run(ctx)
	if err != nil {
		logger.Fatal("bot ran into an error while running", zap.NamedError("RuntimeError", err))
	}
	fmt.Println()
	logger.Info("🤖Done.")
}
