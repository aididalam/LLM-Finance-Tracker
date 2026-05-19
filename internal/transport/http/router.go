package httpapi

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/handler"
	apimiddleware "github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/web"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func NewRouter(
	cfg *config.Config,
	db *sqlx.DB,
	expenseSvc *service.ExpenseService,
	categorySvc *service.CategoryService,
	subcategorySvc *service.SubcategoryService,
	budgetSvc *service.BudgetService,
	settingSvc *service.SettingService,
	walletSvc *service.WalletService,
	llmProvider llm.Provider,
	llmUsage *mysql.LLMUsageRepository,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.Logger)

	// Health (public, no auth)
	health := handler.NewHealthHandler(db)
	r.Get("/health", health.Live)
	r.Get("/health/ready", health.Ready)

	// Auth routes (public, no JWT required)
	authHandler := handler.NewAuthHandler(cfg, db)
	r.Post("/auth/telegram", authHandler.TelegramLogin)
	r.Get("/auth/bot-name", authHandler.BotName)

	// API routes: inject the bootstrapped default user until request-scoped auth is enabled.
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apimiddleware.DefaultUser)

		expenseHandler := handler.NewExpenseHandler(expenseSvc, subcategorySvc).WithLLM(llmProvider, categorySvc, llmUsage)
		r.Get("/expenses", expenseHandler.List)
		r.Post("/expenses", expenseHandler.Create)
		r.Post("/expenses/llm/parse", expenseHandler.LLMParse)
		r.Delete("/expenses/bulk", expenseHandler.BulkDelete)
		r.Put("/expenses/{id}", expenseHandler.Update)
		r.Delete("/expenses/{id}", expenseHandler.Delete)
		r.Get("/expenses/export", handler.NewExportHandler(expenseSvc).CSV)

		categoryHandler := handler.NewCategoryHandler(categorySvc, subcategorySvc)
		r.Get("/categories", categoryHandler.List)
		r.Get("/categories/{id}/subcategories", categoryHandler.ListSubcategories)
		r.Put("/categories/{id}", categoryHandler.Update)
		r.Delete("/categories/{id}", categoryHandler.Delete)

		budgetHandler := handler.NewBudgetHandler(budgetSvc)
		r.Get("/budgets", budgetHandler.List)
		r.Put("/budgets", budgetHandler.Upsert)

		dashboardHandler := handler.NewDashboardHandler(expenseSvc, walletSvc)
		r.Get("/dashboard/overview", dashboardHandler.Overview)
		r.Get("/dashboard/trend", dashboardHandler.Trend)

		settingHandler := handler.NewSettingHandler(settingSvc)
		r.Get("/settings", settingHandler.GetAll)
		r.Put("/settings", settingHandler.Update)

		accountHandler := handler.NewAccountHandler(walletSvc)
		r.Get("/accounts", accountHandler.List)
		r.Post("/accounts", accountHandler.Create)
		r.Put("/accounts/{id}", accountHandler.Update)
		r.Delete("/accounts/{id}", accountHandler.Deactivate)
		r.Post("/accounts/transfer", accountHandler.Transfer)
		r.Post("/accounts/{id}/initial-balance", accountHandler.AddInitialBalance)
		r.Post("/accounts/{id}/income", accountHandler.AddIncome)
		r.Get("/accounts/{id}/debit-cards", accountHandler.ListDebitCards)
		r.Post("/accounts/{id}/debit-cards", accountHandler.AddDebitCard)

		receiptHandler := handler.NewReceiptHandler(llmProvider, categorySvc, llmUsage, settingSvc)
		r.Post("/receipt/parse", receiptHandler.Parse)

		chatHandler := handler.NewChatHandler(llmProvider, categorySvc, expenseSvc, walletSvc, db, llmUsage)
		r.Post("/chat", chatHandler.Handle)
		r.Post("/chat/confirm", chatHandler.Confirm)
		r.Get("/chat/history", chatHandler.GetHistory)
		r.Delete("/chat/history", chatHandler.ClearHistory)
	})

	// Web UI: serve compiled React assets and fall back to index.html for SPA routes.
	webFiles, _ := fs.Sub(web.Files, "files")
	r.Handle("/*", spaHandler(webFiles))

	return r
}

func spaHandler(files fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") || strings.HasPrefix(r.URL.Path, "/health") {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(files, name); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, files, "index.html")
	})
}
