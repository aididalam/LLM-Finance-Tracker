package mysql

import (
	"fmt"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func Connect(cfg *config.Config) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info().Msg("database connected")

	return db
}
