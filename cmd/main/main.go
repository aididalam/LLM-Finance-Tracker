package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/aididalam/llmexpensetracker/internal/app"
	"github.com/aididalam/llmexpensetracker/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to intialize app")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("Application error")
	}

	fmt.Println(cfg.AppEnv)
}
