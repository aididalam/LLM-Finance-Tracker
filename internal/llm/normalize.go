package llm

import (
	"math"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/llm/normalize"
)

func (p *ParsedExpense) Normalize(defaultCurrency ...string) {
	if p == nil {
		return
	}

	dc := "USD"
	if len(defaultCurrency) > 0 && defaultCurrency[0] != "" {
		dc = defaultCurrency[0]
	}

	p.Currency = normalize.Currency(p.Currency, dc)
	p.TransactionType = NormalizeTransactionType(p.TransactionType)
	p.PaymentMethod = normalize.PaymentMethod(p.PaymentMethod)
	p.MovementType = normalizeMovementType(p.MovementType)
	p.ExpenseDate = normalizeDate(p.ExpenseDate)
	p.QueryType = normalizeQueryType(p.QueryType)

	p.Description = strings.TrimSpace(p.Description)
	p.Merchant = strings.TrimSpace(p.Merchant)
	p.Category = cleanCategoryName(p.Category)
	p.Subcategory = strings.TrimSpace(p.Subcategory)
	p.QueryCategory = cleanCategoryName(p.QueryCategory)
	p.QueryFrom = normalizeOptionalDate(p.QueryFrom)
	p.QueryTo = normalizeOptionalDate(p.QueryTo)
	p.NotExpenseReply = strings.TrimSpace(p.NotExpenseReply)
	p.Counterparty = strings.TrimSpace(p.Counterparty)

	if p.Amount < 0 {
		p.Amount = math.Abs(p.Amount)
	}
	if !p.IsExpense && p.MovementType == "" && p.QueryType == "" && p.NotExpenseReply == "" && p.Amount > 0 && (p.Category != "" || p.Description != "") {
		p.IsExpense = true
	}

	if p.MovementType == "" && p.IsExpense {
		p.MovementType = "transaction"
	}
	switch p.MovementType {
	case "credit_card_payment":
		p.IsExpense = false
		if p.Amount > 0 {
			p.NeedsContext = false
		}
		if p.PaymentMethod == "" || p.PaymentMethod == "card_unknown" || p.PaymentMethod == "credit_card" {
			p.PaymentMethod = "bank"
		}
	case "loan_received", "loan_repayment":
		p.IsExpense = false
		if p.Amount > 0 {
			p.NeedsContext = false
		}
		if p.PaymentMethod == "" || p.PaymentMethod == "card_unknown" {
			p.PaymentMethod = "cash"
		}
	}
	if p.PaymentMethod == "card_unknown" && p.MovementType == "transaction" {
		p.MovementType = "ambiguous_card"
		p.NeedsContext = true
		p.IsExpense = false
		if p.NotExpenseReply == "" {
			p.NotExpenseReply = "This says card. Was it debit or credit?"
		}
	}

	if p.IsExpense {
		if p.Category == "" {
			p.Category = "Other"
		}
		if p.Subcategory == "" {
			p.Subcategory = p.Category
		}
		if p.Description == "" {
			p.Description = p.Category + " transaction"
		}
	}
}

func NormalizeTransactionType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "income", "earning", "earnings", "revenue", "salary", "credit", "received", "receive":
		return "income"
	default:
		return "expense"
	}
}

func normalizeMovementType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "transaction", "expense", "income":
		return ""
	case "credit_card_payment", "credit card payment", "card_payment", "cc_payment", "bill_payment":
		return "credit_card_payment"
	case "loan_received", "loan received", "borrowed", "borrow":
		return "loan_received"
	case "loan_repayment", "loan repayment", "repayment", "payback", "paid_back":
		return "loan_repayment"
	case "ambiguous_card", "card_unknown", "unknown_card":
		return "ambiguous_card"
	default:
		return ""
	}
}

func normalizeQueryType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "spending_sum", "expense_sum", "expenses_sum":
		return "spending_sum"
	case "spending_list", "expense_list", "expenses_list":
		return "spending_list"
	case "income_sum", "earning_sum", "earnings_sum":
		return "income_sum"
	case "income_list", "earning_list", "earnings_list":
		return "income_list"
	case "delete_search", "delete":
		return "delete_search"
	default:
		return ""
	}
}

func normalizeDate(value string) string {
	if d := parseDate(value); !d.IsZero() {
		return d.Format("2006-01-02")
	}
	return time.Now().Format("2006-01-02")
}

func normalizeOptionalDate(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if d := parseDate(value); !d.IsZero() {
		return d.Format("2006-01-02")
	}
	return ""
}

func parseDate(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", "02-01-2006", "02/01/2006"} {
		if d, err := time.Parse(layout, value); err == nil {
			return d
		}
	}
	return time.Time{}
}

func cleanCategoryName(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.Index(value, ":"); idx >= 0 {
		value = strings.TrimSpace(value[idx+1:])
	}
	return value
}
