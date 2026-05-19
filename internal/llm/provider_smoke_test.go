package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestProviderSmokeMultilingualInputs(t *testing.T) {
	if os.Getenv("LLM_SMOKE") != "1" {
		t.Skip("set LLM_SMOKE=1 to run provider smoke tests")
	}
	_ = godotenv.Load("../../.env")

	providerName := os.Getenv("LLM_PROVIDER")
	var provider Provider
	switch providerName {
	case "anthropic":
		provider = NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), getenv("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001"))
	case "openai":
		provider = NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), getenv("OPENAI_MODEL", "gpt-4o-mini"))
	default:
		t.Fatalf("unsupported or missing LLM_PROVIDER %q", providerName)
	}

	categories := []string{
		"expense: Food", "expense: Transport", "expense: Shopping", "expense: Bills", "expense: Other",
		"income: Salary", "income: Freelance", "income: Business", "income: Gift", "income: Refund", "income: Other",
	}

	cases := []struct {
		name      string
		input     string
		wantKind  string
		wantQuery string
	}{
		{name: "english expense", input: "Spent 250 taka on lunch at Sultan's Dine", wantKind: "expense"},
		{name: "banglish expense", input: "ajke office jawar jonno rickshaw e 120 taka dilam", wantKind: "expense"},
		{name: "bangla expense", input: "আজকে বাজারে ৭৫০ টাকা খরচ করেছি", wantKind: "expense"},
		{name: "casual mixed single transaction", input: "bro ajke lunch e 220 taka, btw weather ta onek kharap", wantKind: "expense"},
		{name: "english income", input: "Got salary 50000 BDT today", wantKind: "income"},
		{name: "banglish income", input: "client theke freelance payment 15000 taka peyechi", wantKind: "income"},
		{name: "multiple transactions unsafe", input: "salary 50000 peyechi and lunch e 300 taka spend korlam", wantKind: "none"},
		{name: "mixed query expense", input: "ei month food e koto spend korlam?", wantQuery: "spending_sum"},
		{name: "mixed query income", input: "amar ei mashe income koto holo?", wantQuery: "income_sum"},
		{name: "unknown greeting", input: "hello bhai, ki obostha?", wantKind: "none"},
		{name: "random unknown", input: "asdf qwer blue tomato 44 maybe later", wantKind: "none"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			parsed, _, err := provider.ParseExpense(ctx, tc.input, categories, "BDT")
			if err != nil {
				t.Fatalf("ParseExpense error: %v", err)
			}
			parsed.Normalize("BDT")

			t.Logf("input=%q parsed=%+v", tc.input, *parsed)

			if tc.wantQuery != "" {
				if parsed.IsExpense {
					t.Fatalf("got transaction, want query %s", tc.wantQuery)
				}
				if parsed.QueryType != tc.wantQuery {
					t.Fatalf("query_type=%q, want %q", parsed.QueryType, tc.wantQuery)
				}
				return
			}

			switch tc.wantKind {
			case "expense", "income":
				if !parsed.IsExpense {
					t.Fatalf("got non-transaction, want %s", tc.wantKind)
				}
				if parsed.TransactionType != tc.wantKind {
					t.Fatalf("transaction_type=%q, want %q", parsed.TransactionType, tc.wantKind)
				}
				if parsed.Amount <= 0 {
					t.Fatalf("amount=%v, want positive", parsed.Amount)
				}
				if parsed.Category == "" || parsed.Subcategory == "" || parsed.Description == "" {
					t.Fatalf("category/subcategory/description should be populated: %+v", *parsed)
				}
			case "none":
				if parsed.IsExpense || parsed.QueryType != "" {
					t.Fatalf("got transaction/query for unknown input: %+v", *parsed)
				}
				if parsed.NotExpenseReply == "" {
					t.Fatalf("not_expense_reply should be populated for unknown input")
				}
			}
		})
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
