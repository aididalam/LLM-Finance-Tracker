package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/jmoiron/sqlx"
)

// ChatHandler serves the conversational AI endpoint used by the web chat panel.
type ChatHandler struct {
	llmProvider llm.Provider
	categorySvc *service.CategoryService
	expenseSvc  *service.ExpenseService
	walletSvc   *service.WalletService
	db          *sqlx.DB
	llmUsage    *mysql.LLMUsageRepository
}

// NewChatHandler creates a ChatHandler.
func NewChatHandler(llmProvider llm.Provider, categorySvc *service.CategoryService, expenseSvc *service.ExpenseService, walletSvc *service.WalletService, db *sqlx.DB, llmUsage *mysql.LLMUsageRepository) *ChatHandler {
	return &ChatHandler{
		llmProvider: llmProvider,
		categorySvc: categorySvc,
		expenseSvc:  expenseSvc,
		walletSvc:   walletSvc,
		db:          db,
		llmUsage:    llmUsage,
	}
}

type chatRequest struct {
	Messages []llm.ChatMessage `json:"messages"`
}

// ChatReply is the JSON payload returned by POST /api/v1/chat.
type ChatReply struct {
	Action        string                        `json:"action"` // "confirm" | "clarify" | "answer" | "chat" | "delete_select"
	Reply         string                        `json:"reply"`
	Expense       *llm.ParsedExpense            `json:"expense,omitempty"`        // action="confirm"
	Expenses      []*domain.ExpenseWithCategory `json:"expenses,omitempty"`       // action="delete_select"
	WalletOptions []*domain.WalletWithBalance   `json:"wallet_options,omitempty"` // action="confirm"
}

const maxChatHistory = 10

// Handle processes a multi-turn chat message.
//
//	POST /api/v1/chat
func (h *ChatHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.llmProvider == nil {
		response.ResError(w, "LLM not configured", http.StatusServiceUnavailable)
		return
	}

	userID := middleware.CurrentUserID(r.Context())
	categorySvc := h.categorySvc.ForUser(userID)
	expenseSvc := h.expenseSvc.ForUser(userID)
	var walletSvc *service.WalletService
	if h.walletSvc != nil {
		walletSvc = h.walletSvc.ForUser(userID)
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Messages) == 0 {
		response.ResError(w, "messages array is required")
		return
	}

	msgs := req.Messages
	if len(msgs) > maxChatHistory {
		msgs = msgs[len(msgs)-maxChatHistory:]
	}

	userMsg := msgs[len(msgs)-1]
	if userMsg.Role == "user" {
		h.saveMessageForUser(r.Context(), userID, "user", userMsg.Content)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	categories, err := categorySvc.Names()
	if err != nil {
		response.ResError(w, "failed to load categories", http.StatusInternalServerError)
		return
	}

	proc := service.NewChatProcessor(h.llmProvider, categorySvc, expenseSvc, walletSvc)
	decision, usage, err := proc.ProcessMessages(ctx, msgs, categories)
	if err != nil {
		response.ResError(w, "AI error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if usage != nil && h.llmUsage != nil {
		_ = h.llmUsage.ForUser(userID).Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "web_chat")
	}

	reply := ChatReply{
		Action:   decision.Action,
		Reply:    decision.Reply,
		Expense:  decision.Expense,
		Expenses: decision.Expenses,
	}

	if decision.Action == service.ChatActionConfirm && decision.Expense != nil && walletSvc != nil {
		if ws, err := walletSvc.ListWithBalances(); err == nil {
			reply.WalletOptions = service.MatchWallets(decision.Expense.PaymentMethod, ws)
		}
	}

	h.saveMessageForUser(r.Context(), userID, "assistant", reply.Reply)
	response.ResSuccess(w, reply)
}

// GetHistory returns saved chat messages.
//
//	GET /api/v1/chat/history
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	type row struct {
		Role    string `db:"role"    json:"role"`
		Content string `db:"content" json:"content"`
	}
	var msgs []row
	if err := h.db.SelectContext(r.Context(), &msgs,
		`SELECT role, content FROM chat_messages WHERE user_id = ? ORDER BY id ASC LIMIT 200`,
		userID); err != nil {
		response.ResError(w, "db error", http.StatusInternalServerError)
		return
	}
	if msgs == nil {
		msgs = []row{}
	}
	response.ResSuccess(w, msgs)
}

// ClearHistory deletes all saved chat messages.
//
//	DELETE /api/v1/chat/history
func (h *ChatHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	if _, err := h.db.ExecContext(r.Context(), `DELETE FROM chat_messages WHERE user_id = ?`, userID); err != nil {
		response.ResError(w, "db error", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, map[string]bool{"ok": true})
}

// Confirm saves a confirmed expense from the chat panel.
//
//	POST /api/v1/chat/confirm
func (h *ChatHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	expenseSvc := h.expenseSvc.ForUser(userID)

	var body struct {
		Parsed                *llm.ParsedExpense `json:"expense"`
		WalletID              string             `json:"wallet_id"`
		WalletBankDebitCardID *string            `json:"wallet_bank_debit_card_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Parsed == nil {
		response.ResError(w, "expense is required")
		return
	}
	if body.WalletID == "" && h.walletSvc != nil {
		wid, err := h.walletSvc.ForUser(userID).DefaultWalletID()
		if err == nil {
			body.WalletID = wid
		}
	}

	var catID *string
	if h.categorySvc != nil {
		cat, err := h.categorySvc.ForUser(userID).FindOrCreate(body.Parsed.Category)
		if err == nil && cat != nil {
			catID = &cat.ID
		}
	}

	expense, isIncome, err := expenseSvc.CreateFromParsed(body.Parsed, catID, body.WalletID, "text", body.WalletBankDebitCardID)
	if err != nil {
		response.ResError(w, "failed to save expense", http.StatusInternalServerError)
		return
	}
	if isIncome {
		response.ResSuccess(w, map[string]any{"saved": true, "is_income": true})
		return
	}
	response.ResSuccess(w, map[string]any{"saved": true, "expense": expense})
}

func (h *ChatHandler) saveMessageForUser(ctx context.Context, userID, role, content string) {
	h.db.ExecContext(ctx, `INSERT INTO chat_messages (user_id, role, content) VALUES (?, ?, ?)`, userID, role, content)
}
