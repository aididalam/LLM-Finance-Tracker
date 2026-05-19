package llm

import (
	"testing"
	"time"
)

func TestParsedExpenseNormalizeTransaction(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:       true,
		Amount:          -120,
		Currency:        " taka ",
		Category:        "income: Salary",
		TransactionType: "Income",
		PaymentMethod:   "cc",
		ExpenseDate:     "2026/05/17",
	}

	p.Normalize()

	if p.TransactionType != "income" {
		t.Fatalf("transaction type = %q, want income", p.TransactionType)
	}
	if p.Amount != 120 {
		t.Fatalf("amount = %v, want 120", p.Amount)
	}
	if p.Currency != "BDT" {
		t.Fatalf("currency = %q, want BDT", p.Currency)
	}
	if p.Category != "Salary" {
		t.Fatalf("category = %q, want Salary", p.Category)
	}
	if p.Subcategory != "Salary" {
		t.Fatalf("subcategory = %q, want Salary", p.Subcategory)
	}
	if p.PaymentMethod != "credit_card" {
		t.Fatalf("payment method = %q, want credit_card", p.PaymentMethod)
	}
	if p.ExpenseDate != "2026-05-17" {
		t.Fatalf("expense date = %q, want 2026-05-17", p.ExpenseDate)
	}
}

func TestParsedExpenseNormalizeCompactCardWords(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:       true,
		Amount:          120,
		Currency:        "BDT",
		Category:        "Food",
		TransactionType: "expense",
		PaymentMethod:   "creditcard",
		ExpenseDate:     "2026-05-18",
	}

	p.Normalize()

	if p.PaymentMethod != "credit_card" {
		t.Fatalf("payment method = %q, want credit_card", p.PaymentMethod)
	}
}

func TestParsedExpenseNormalizeGenericCardNeedsContext(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:       true,
		Amount:          120,
		Currency:        "BDT",
		Category:        "Food",
		TransactionType: "expense",
		PaymentMethod:   "card",
		ExpenseDate:     "2026-05-17",
	}

	p.Normalize()

	if p.IsExpense {
		t.Fatal("generic card should not be saved without debit/credit clarification")
	}
	if !p.NeedsContext {
		t.Fatal("generic card should ask for context")
	}
	if p.PaymentMethod != "card_unknown" {
		t.Fatalf("payment method = %q, want card_unknown", p.PaymentMethod)
	}
	if p.MovementType != "ambiguous_card" {
		t.Fatalf("movement type = %q, want ambiguous_card", p.MovementType)
	}
}

func TestParsedExpenseNormalizeCreditCardPaymentDoesNotUseUnknownCard(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:     false,
		Amount:        200,
		Currency:      "BDT",
		Description:   "Credit card bill payment",
		MovementType:  "credit_card_payment",
		PaymentMethod: "card_unknown",
		ExpenseDate:   "2026-05-18",
	}

	p.Normalize()

	if p.IsExpense {
		t.Fatal("credit card payment should not be expense")
	}
	if p.PaymentMethod != "bank" {
		t.Fatalf("payment method = %q, want bank", p.PaymentMethod)
	}
	if p.MovementType != "credit_card_payment" {
		t.Fatalf("movement type = %q, want credit_card_payment", p.MovementType)
	}
}

func TestParsedExpenseNormalizeUnknownTransaction(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:     true,
		Currency:      "something weird",
		PaymentMethod: "bkash",
		ExpenseDate:   "not a date",
	}

	p.Normalize()

	if p.TransactionType != "expense" {
		t.Fatalf("transaction type = %q, want expense", p.TransactionType)
	}
	if p.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", p.Currency)
	}
	if p.Category != "Other" {
		t.Fatalf("category = %q, want Other", p.Category)
	}
	if p.PaymentMethod != "mobile_wallet" {
		t.Fatalf("payment method = %q, want mobile_wallet", p.PaymentMethod)
	}
	if _, err := time.Parse("2006-01-02", p.ExpenseDate); err != nil {
		t.Fatalf("expense date = %q, want valid YYYY-MM-DD", p.ExpenseDate)
	}
}

func TestParsedExpenseNormalizeCorrectsIncomeBoolean(t *testing.T) {
	p := &ParsedExpense{
		IsExpense:       false,
		Amount:          15000,
		Currency:        "BDT",
		Description:     "Freelance payment from client",
		Category:        "Freelance",
		Subcategory:     "Project Payment",
		TransactionType: "income",
	}

	p.Normalize()

	if !p.IsExpense {
		t.Fatal("income transaction with amount/category should be saveable even if provider set is_expense=false")
	}
	if p.TransactionType != "income" {
		t.Fatalf("transaction type = %q, want income", p.TransactionType)
	}
}

func TestParsedExpenseNormalizeQueries(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"expense_sum", "spending_sum"},
		{"expenses_list", "spending_list"},
		{"earning_sum", "income_sum"},
		{"earnings_list", "income_list"},
		{"unknown", ""},
	}

	for _, tc := range cases {
		p := &ParsedExpense{QueryType: tc.in, QueryFrom: "17/05/2026", QueryTo: "bad"}
		p.Normalize()
		if p.QueryType != tc.want {
			t.Fatalf("query type %q normalized to %q, want %q", tc.in, p.QueryType, tc.want)
		}
		if p.QueryFrom != "2026-05-17" {
			t.Fatalf("query from = %q, want 2026-05-17", p.QueryFrom)
		}
		if p.QueryTo != "" {
			t.Fatalf("query to = %q, want empty for bad date", p.QueryTo)
		}
	}
}
