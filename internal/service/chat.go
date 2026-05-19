package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
)

var amountPattern = regexp.MustCompile(`(?i)(?:৳|bdt|tk|taka)?\s*([0-9][0-9,]*(?:\.[0-9]+)?)`)

// IsAffirmative returns true for common yes-like replies (English + Bengali).
func IsAffirmative(text string) bool {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "yes", "y", "yeah", "yep", "ha", "haan", "hmm", "ok", "জি", "হ্যাঁ":
		return true
	}
	return false
}

// IsNegative returns true for common no-like replies (English + Bengali).
func IsNegative(text string) bool {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "no", "n", "nope", "na", "nah", "না":
		return true
	}
	return false
}

// ParseAmount extracts the first numeric amount from text, returning 0 if none found.
func ParseAmount(text string) float64 {
	match := amountPattern.FindStringSubmatch(strings.ToLower(text))
	if len(match) < 2 {
		return 0
	}
	val, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
	if err != nil {
		return 0
	}
	return val
}

const (
	ChatActionChat         = "chat"
	ChatActionConfirm      = "confirm"
	ChatActionClarify      = "clarify"
	ChatActionAnswer       = "answer"
	ChatActionDeleteSelect = "delete_select"
)

type ChatDecision struct {
	Action   string
	Reply    string
	Expense  *llm.ParsedExpense
	Expenses []*domain.ExpenseWithCategory
}

type ChatProcessor struct {
	llmProvider llm.Provider
	categorySvc *CategoryService
	expenseSvc  *ExpenseService
	walletSvc   *WalletService
}

func NewChatProcessor(llmProvider llm.Provider, categorySvc *CategoryService, expenseSvc *ExpenseService, walletSvc *WalletService) *ChatProcessor {
	return &ChatProcessor{
		llmProvider: llmProvider,
		categorySvc: categorySvc,
		expenseSvc:  expenseSvc,
		walletSvc:   walletSvc,
	}
}

func (p *ChatProcessor) ProcessMessages(ctx context.Context, messages []llm.ChatMessage, categories []string) (ChatDecision, *llm.Usage, error) {
	if p.llmProvider == nil {
		return ChatDecision{}, nil, fmt.Errorf("LLM not configured")
	}
	parsed, usage, err := p.llmProvider.Chat(ctx, messages, categories, "")
	if err != nil {
		return ChatDecision{}, usage, err
	}
	return p.ProcessParsed(parsed), usage, nil
}

func (p *ChatProcessor) ProcessParsed(parsed *llm.ParsedExpense) ChatDecision {
	if parsed == nil {
		return ChatDecision{Action: ChatActionChat, Reply: "I couldn't understand that. Please try again."}
	}
	parsed.Normalize()

	switch {
	case parsed.IsExpense:
		return ChatDecision{
			Action:  ChatActionConfirm,
			Reply:   BuildChatConfirmReply(parsed),
			Expense: parsed,
		}

	case parsed.NeedsContext:
		reply := parsed.NotExpenseReply
		if reply == "" {
			reply = "Could you provide more details?"
		}
		return ChatDecision{Action: ChatActionClarify, Reply: reply}

	case parsed.QueryType == "delete_search":
		expenses, reply := p.runDeleteSearch(parsed)
		if len(expenses) == 0 {
			return ChatDecision{Action: ChatActionChat, Reply: reply}
		}
		return ChatDecision{Action: ChatActionDeleteSelect, Reply: reply, Expenses: expenses}

	case parsed.QueryType != "":
		return ChatDecision{Action: ChatActionAnswer, Reply: p.runQuery(parsed)}

	default:
		reply := parsed.NotExpenseReply
		if reply == "" {
			reply = "I'm here to help you track your finances. Tell me about a transaction or ask about your spending!"
		}
		return ChatDecision{Action: ChatActionChat, Reply: reply}
	}
}

func (p *ChatProcessor) runDeleteSearch(parsed *llm.ParsedExpense) ([]*domain.ExpenseWithCategory, string) {
	if p.expenseSvc == nil {
		return nil, "Transaction search is not configured."
	}

	from, to := chatQueryDates(parsed)
	filters := domain.ExpenseFilters{From: &from, To: &to, Limit: 15}
	if parsed.DeleteKeywords != "" {
		filters.Search = &parsed.DeleteKeywords
	}
	if parsed.QueryCategory != "" && p.categorySvc != nil {
		if cat, _ := p.categorySvc.FindByName(parsed.QueryCategory); cat != nil {
			filters.CategoryID = &cat.ID
		}
	}

	expenses, err := p.expenseSvc.FindAllWithCategory(filters)
	if err != nil || len(expenses) == 0 {
		return nil, "No matching expenses found for that description."
	}
	return expenses, fmt.Sprintf("Found %d matching expense(s). Tap one to delete:", len(expenses))
}

func (p *ChatProcessor) runQuery(parsed *llm.ParsedExpense) string {
	if p.expenseSvc == nil {
		return "Transaction search is not configured."
	}

	from, to := chatQueryDates(parsed)
	filters := domain.ExpenseFilters{From: &from, To: &to, Limit: 500}
	if parsed.QueryCategory != "" && p.categorySvc != nil {
		if cat, _ := p.categorySvc.FindByName(parsed.QueryCategory); cat != nil {
			filters.CategoryID = &cat.ID
		}
	}

	expenses, err := p.expenseSvc.FindAllWithCategory(filters)
	if err != nil {
		return "Sorry, I couldn't fetch your transactions right now."
	}

	var total float64
	for _, e := range expenses {
		total += e.Amount
	}

	catLabel := ""
	if parsed.QueryCategory != "" {
		catLabel = " on " + parsed.QueryCategory
	}

	if parsed.QueryType == "spending_sum" {
		return fmt.Sprintf("You spent %.2f%s between %s and %s (%d transactions).",
			total, catLabel, from, to, len(expenses))
	}

	if len(expenses) == 0 {
		return "No transactions found for that period."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Expenses%s (%s to %s):\n", catLabel, from, to))

	limit := len(expenses)
	if limit > 10 {
		limit = 10
	}
	for _, e := range expenses[:limit] {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		sb.WriteString(fmt.Sprintf("• %s — %.2f (%s)\n",
			desc, e.Amount, e.ExpenseDatetime.Format("02 Jan")))
	}
	if len(expenses) > 10 {
		sb.WriteString(fmt.Sprintf("... and %d more", len(expenses)-10))
	}
	return sb.String()
}

// BuildChatConfirmReply returns a confirmation message for a parsed expense.
func BuildChatConfirmReply(p *llm.ParsedExpense) string {
	txLabel := "Expense"
	if IsIncomeType(p.TransactionType) {
		txLabel = "Income"
	}
	cat := p.Category
	if p.Subcategory != "" {
		cat = p.Category + " › " + p.Subcategory
	}
	date := p.ExpenseDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return fmt.Sprintf("%s detected: %.2f for \"%s\" (%s) on %s. Confirm to save?",
		txLabel, p.Amount, p.Description, cat, date)
}

func chatQueryDates(parsed *llm.ParsedExpense) (string, string) {
	now := time.Now()
	from := parsed.QueryFrom
	if from == "" {
		from = fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	}
	to := parsed.QueryTo
	if to == "" {
		to = now.Format("2006-01-02")
	}
	return from, to
}

// MatchWallets filters wallets by the LLM-detected payment method hint.
// Returns only wallets matching the hint, or all active wallets when no hint matches.
func MatchWallets(paymentMethod string, wallets []*domain.WalletWithBalance) []*domain.WalletWithBalance {
	pm := strings.ToLower(paymentMethod)
	var matchType string
	switch {
	case strings.Contains(pm, "cash"):
		matchType = "cash"
	case strings.Contains(pm, "bank") || strings.Contains(pm, "card") || strings.Contains(pm, "transfer"):
		matchType = "bank"
	case strings.Contains(pm, "mfs") || strings.Contains(pm, "mobile"):
		matchType = "mfs"
	}
	if matchType != "" {
		var matched []*domain.WalletWithBalance
		for _, w := range wallets {
			if w.IsActive && w.AccountType == matchType {
				matched = append(matched, w)
			}
		}
		if len(matched) > 0 {
			return matched
		}
	}
	var all []*domain.WalletWithBalance
	for _, w := range wallets {
		if w.IsActive {
			all = append(all, w)
		}
	}
	return all
}
