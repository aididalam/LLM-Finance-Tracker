package main

import (
	"os"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/rs/zerolog/log"
)

func main() {
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	if direction != "up" && direction != "down" {
		log.Fatal().Msgf("usage: migrate [up|down]")
	}

	cfg := config.Load()
	db := mysql.Connect(cfg)
	defer db.Close()

	mysql.RunMigration(cfg, db, direction)
}
