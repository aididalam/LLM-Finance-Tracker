package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	httpapi "github.com/aididalam/llmexpensetracker/internal/transport/api"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Application struct {
	cfg    *config.Config
	db     *sqlx.DB
	server *http.Server
}

type Repositories struct {
}
type Services struct {
}

func New(cfg *config.Config) (*Application, error) {
	//db connect
	db := mysql.Connect(cfg)
	if cfg.AppEnv == "development" {
		mysql.RunMigration(cfg, "up")
	}

	// repository init
	repos := newRepositories(db)

	//service init
	_ = newServices(repos)

	router := httpapi.Router(cfg, db)

	return &Application{
		cfg: cfg,
		db:  db,
		server: &http.Server{
			Addr:         ":" + cfg.AppPort,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	defer a.db.Close()

	serverErr := make(chan error, 1)

	go func() {
		log.Info().Msgf("server starting on :%s", a.cfg.AppPort)

		err := a.server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")

	case err := <-serverErr:
		return fmt.Errorf("server failed: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("shutting down server")

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	log.Info().Msg("server stopped")
	return nil
}

func newRepositories(db *sqlx.DB) Repositories {
	return Repositories{}
}

func newServices(repos Repositories) Services {
	return Services{}
}
