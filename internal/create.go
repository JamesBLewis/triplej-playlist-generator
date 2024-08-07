package internal

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/JamesBLewis/triplej-playlist-generator/internal/config"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/log"
	"github.com/JamesBLewis/triplej-playlist-generator/pkg/telemetry"
)

func RunBot() error {
	ctx := context.Background()

	otelShutdown, err := telemetry.InitTelemetry()
	if err != nil {
		return errors.Wrap(err, "failed to configure OpenTelemetry")
	}
	defer otelShutdown()

	// Add a child span
	ctx, childSpan := otel.Tracer(telemetry.TracerName).Start(ctx, "RunBot")
	defer childSpan.End()

	// Instantiate a new slog logger
	logger := log.NewLogger()

	runtimeErr := createBot(ctx, logger)
	if runtimeErr != nil {
		logger.RuntimeError(ctx, "An error occurred while running the bot", runtimeErr)
		return runtimeErr
	}
	return nil
}

func createBot(ctx context.Context, logger log.Log) error {
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}
	bot := NewBot(cfg, logger)
	err = bot.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "bot ran into an error")
	}
	return nil
}
