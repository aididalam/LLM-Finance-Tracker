package mysql

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
)

func RunMigration(cfg *config.Config, direction string) {
	files, _ := filepath.Glob("migrations/*.sql")

	if len(files) == 0 {
		log.Info().Msg("no migration files found, skipping")
		return
	}

	migDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&multiStatements=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	migDB, err := sql.Open("mysql", migDSN)

	if err != nil {
		log.Fatal().Err(err).Msg("migration db open failed")
	}
	defer migDB.Close()

	driver, err := migmysql.WithInstance(migDB, &migmysql.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("migration driver init failed")
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", cfg.DBName, driver)
	if err != nil {
		log.Fatal().Err(err).Msg("migration init failed")
	}

	switch direction {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	default:
		log.Fatal().Msgf("unknown migration direction: %s", direction)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msgf("migration %s failed", direction)
	}

	log.Info().Msgf("migrations %s applied", direction)

}
