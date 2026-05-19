package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/aididalam/llmexpensetracker/internal/app"
	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}
}
