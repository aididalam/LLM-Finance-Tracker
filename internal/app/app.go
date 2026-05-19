package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	httpapi "github.com/aididalam/llmexpensetracker/internal/transport/http"
	"github.com/aididalam/llmexpensetracker/internal/transport/telegram"
	"github.com/aididalam/llmexpensetracker/internal/worker"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Application struct {
	cfg      *config.Config
	db       *sqlx.DB
	server   *http.Server
	telegram *telegram.Handler
	services services
}

type repositories struct {
	categories    domain.CategoryRepository
	subcategories domain.SubcategoryRepository
	expenses      domain.ExpenseRepository
	budgets       domain.BudgetRepository
	settings      domain.SettingRepository
	wallets       domain.WalletRepository
	llmUsage      *mysql.LLMUsageRepository
}

type services struct {
	categories    *service.CategoryService
	subcategories *service.SubcategoryService
	expenses      *service.ExpenseService
	budgets       *service.BudgetService
	settings      *service.SettingService
	wallets       *service.WalletService
}

func New(cfg *config.Config) (*Application, error) {
	configureLogger(cfg)

	db := mysql.Connect(cfg)
	mysql.RunMigration(cfg, db, "up")

	repos := newRepositories(db)
	services := newServices(repos)

	llmProvider, err := newLLMProvider(cfg)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	log.Info().Msgf("LLM provider: %s (%s)", llmProvider.Name(), llmProvider.Model())

	telegramHandler, err := telegram.NewHandler(
		cfg,
		llmProvider,
		services.categories,
		services.expenses,
		services.budgets,
		services.wallets,
		repos.llmUsage,
		services.settings,
	)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create telegram handler: %w", err)
	}

	router := httpapi.NewRouter(
		cfg,
		db,
		services.expenses,
		services.categories,
		services.subcategories,
		services.budgets,
		services.settings,
		services.wallets,
		llmProvider,
		repos.llmUsage,
	)

	return &Application{
		cfg:      cfg,
		db:       db,
		telegram: telegramHandler,
		services: services,
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

	rootCtx, stopBackground := context.WithCancel(ctx)
	defer stopBackground()

	go a.telegram.StartPolling(rootCtx)

	alerter := worker.NewBudgetAlerter(
		a.telegram.BotAPI(),
		a.telegram.ChatID(),
		a.services.budgets,
		a.services.settings,
	)
	scheduler, err := worker.StartScheduler(a.telegram.BotAPI(), a.telegram.ChatID(), a.services.expenses, alerter)
	if err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Error().Err(err).Msg("scheduler: shutdown failed")
		}
	}()

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Msgf("server starting on :%s", a.cfg.AppPort)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		stopBackground()
		return fmt.Errorf("server failed: %w", err)
	}

	stopBackground()
	log.Info().Msg("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	log.Info().Msg("server stopped")
	return nil
}

func configureLogger(cfg *config.Config) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	if cfg.AppEnv == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func newRepositories(db *sqlx.DB) repositories {
	return repositories{
		categories:    mysql.NewCategoryRepository(db),
		subcategories: mysql.NewSubcategoryRepository(db),
		expenses:      mysql.NewExpenseRepository(db),
		budgets:       mysql.NewBudgetRepository(db),
		settings:      mysql.NewSettingRepository(db),
		wallets:       mysql.NewWalletRepository(db),
		llmUsage:      mysql.NewLLMUsageRepository(db),
	}
}

func newServices(repos repositories) services {
	categorySvc := service.NewCategoryService(repos.categories)
	subcategorySvc := service.NewSubcategoryService(repos.subcategories)
	settingSvc := service.NewSettingService(repos.settings)
	walletSvc := service.NewWalletService(repos.wallets)
	expenseSvc := service.NewExpenseService(repos.expenses, subcategorySvc).
		WithWallets(walletSvc)

	return services{
		categories:    categorySvc,
		subcategories: subcategorySvc,
		expenses:      expenseSvc,
		budgets:       service.NewBudgetService(repos.budgets, repos.expenses),
		settings:      settingSvc,
		wallets:       walletSvc,
	}
}

func newLLMProvider(cfg *config.Config) (llm.Provider, error) {
	switch cfg.LLMProvider {
	case "anthropic":
		return llm.NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.AnthropicModel), nil
	case "openai":
		return llm.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel), nil
	default:
		return nil, fmt.Errorf("unknown LLM_PROVIDER %q; use 'anthropic' or 'openai'", cfg.LLMProvider)
	}
}
