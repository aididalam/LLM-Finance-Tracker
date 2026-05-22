package main

import (
	"os"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/rs/zerolog/log"
)

func main() {
	direct := "up"

	if len(os.Args) > 1 {
		direct = os.Args[1]
	}

	if direct != "up" && direct != "down" {
		log.Fatal().Msgf("invalid migration direction: %s", direct)
	}

	cfg := config.Load()
	mysql.RunMigration(cfg, direct)
}
