package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ReceiptItem is one line item from a multi-item receipt.
type ReceiptItem struct {
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Category        string  `json:"category"`
	Subcategory     string  `json:"subcategory"`
	TransactionType string  `json:"transaction_type"`
	PaymentMethod   string  `json:"payment_method"`
	MovementType    string  `json:"movement_type"`
	Counterparty    string  `json:"counterparty"`
	ExpenseDate     string  `json:"expense_date"`
	Merchant        string  `json:"merchant"`
}

// ToParsed converts a receipt line item to a ParsedExpense for saving.
func (item *ReceiptItem) ToParsed(defaultCurrency ...string) *ParsedExpense {
	p := &ParsedExpense{
		IsExpense:       true,
		Amount:          item.Amount,
		Currency:        item.Currency,
		Description:     item.Description,
		Merchant:        item.Merchant,
		Category:        item.Category,
		Subcategory:     item.Subcategory,
		TransactionType: item.TransactionType,
		PaymentMethod:   item.PaymentMethod,
		MovementType:    item.MovementType,
		Counterparty:    item.Counterparty,
		ExpenseDate:     item.ExpenseDate,
	}
	p.Normalize(defaultCurrency...)
	return p
}

type ParsedExpense struct {
	IsExpense       bool    `json:"is_expense"`
	NeedsContext    bool    `json:"needs_context"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Description     string  `json:"description"`
	Merchant        string  `json:"merchant"`
	Category        string  `json:"category"`
	Subcategory     string  `json:"subcategory"`
	TransactionType string  `json:"transaction_type"`
	PaymentMethod   string  `json:"payment_method"`
	MovementType    string  `json:"movement_type"`
	Counterparty    string  `json:"counterparty"`
	ExpenseDate     string  `json:"expense_date"`
	NotExpenseReply string  `json:"not_expense_reply"`
	// Query fields — populated when is_expense=false and message is a spending/income query
	QueryType      string `json:"query_type"`      // "spending_sum" | "spending_list" | "income_sum" | "income_list" | "delete_search" | ""
	QueryCategory  string `json:"query_category"`  // category name if mentioned
	QueryFrom      string `json:"query_from"`      // YYYY-MM-DD
	QueryTo        string `json:"query_to"`        // YYYY-MM-DD
	DeleteKeywords string `json:"delete_keywords"` // keywords for delete_search
	// Multi-item receipt — set by ParseReceipt when there are 2+ line items
	Items []ReceiptItem `json:"items,omitempty"`
}

// ChatMessage is a single message in a multi-turn conversation.
type ChatMessage struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens int
	OutputTokens int
}

type Provider interface {
	ParseExpense(ctx context.Context, message string, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error)
	// Chat performs a multi-turn conversation. messages is the full history (user+assistant alternating).
	Chat(ctx context.Context, messages []ChatMessage, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error)
	// ParseReceipt extracts expense details from a receipt image (raw bytes).
	ParseReceipt(ctx context.Context, imageData []byte, mediaType string, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error)
	Name() string
	Model() string
}

func buildSystemPrompt(categories []string, defaultCurrency string) string {
	today := time.Now().Format("2006-01-02")
	thisMonthStart := time.Now().Format("2006-01") + "-01"
	cats := "none yet"
	if len(categories) > 0 {
		cats = strings.Join(categories, ", ")
	}
	return fmt.Sprintf(`You are a personal finance assistant. Analyze the message and respond with a single JSON object.
Today is %s. Existing categories: %s.

Return ONLY a valid JSON object — no markdown, no extra text.

Expense/income: {"is_expense":true,"amount":200.00,"currency":"%s","description":"English description","merchant":"","category":"Transport","subcategory":"Taxi","transaction_type":"expense","payment_method":"cash","movement_type":"transaction","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"","query_category":"","query_from":"","query_to":""}
Credit card bill payment: {"is_expense":false,"amount":10000.00,"currency":"%s","description":"Credit card bill payment","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"bank","movement_type":"credit_card_payment","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"","query_category":"","query_from":"","query_to":""}
Loan received/repayment: {"is_expense":false,"amount":5000.00,"currency":"%s","description":"Loan repayment","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"loan_repayment","counterparty":"Rafi","expense_date":"%s","not_expense_reply":"","query_type":"","query_category":"","query_from":"","query_to":""}
Spending/income query: {"is_expense":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"spending_sum","query_category":"Food","query_from":"%s","query_to":"%s"}
Non-transaction: {"is_expense":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"Reply in same language","query_type":"","query_category":"","query_from":"","query_to":""}

Rules:
- is_expense=true for BOTH expenses and income that should be saved; false for queries, greetings, unclear messages.
- If one clear transaction is mixed with casual text, parse the transaction and ignore the rest.
- If multiple distinct transactions in one message, set is_expense=false and ask user to send one at a time.
- transaction_type: "expense" for outgoing money; "income" for salary, freelance, gifts, refunds, any money coming in.
- currency: default "%s". Detect local currency symbols/words and map to ISO codes.
- payment_method: "cash", "bank", "mobile_wallet", "debit_card", "credit_card", "card_unknown", or "other". Generic "card" without debit/credit → "card_unknown".
- movement_type: "transaction" for ordinary expense/income; "credit_card_payment" for paying a credit card bill; "loan_received" for borrowing; "loan_repayment" for repaying; "ambiguous_card" when only generic "card" is mentioned.
- Credit card purchase: payment_method="credit_card", movement_type="transaction".
- Credit card bill payment (bill/due/minimum): is_expense=false, movement_type="credit_card_payment", payment_method="bank" or "cash".
- If paid by card but no purchase description, set needs_context=true and ask what was bought.
- Loan: borrowed/received → loan_received. repaid/paid back → loan_repayment. Missing amount → needs_context=true.
- expense_date: today (%s) if not mentioned; YYYY-MM-DD.
- category: match existing category or create a short new one.
- subcategory: specific label (Lunch, Taxi, Monthly Salary). Always fill.
- description: always in English.
- not_expense_reply: reply in the same language/script the user used.
- query_type: "spending_sum"/"spending_list" for expense queries; "income_sum"/"income_list" for income queries.
- query_from/query_to: current month by default.`,
		today, cats,
		defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency, today, thisMonthStart, today,
		defaultCurrency, today,
		defaultCurrency,
		today)
}

// buildReceiptSystemPrompt returns the system prompt for vision-based receipt parsing.
// It instructs the LLM to return multiple items when the receipt has 2+ line items.
func buildReceiptSystemPrompt(categories []string, defaultCurrency string) string {
	today := time.Now().Format("2006-01-02")
	cats := "none yet"
	if len(categories) > 0 {
		cats = strings.Join(categories, ", ")
	}
	return fmt.Sprintf(`You are a receipt parser. Extract expense details from the receipt image or document and respond ONLY with a valid JSON object — no markdown, no extra text.
Today is %s. Categories: %s.

Multiple line items (2+): {"is_expense":true,"merchant":"Shop Name","expense_date":"%s","items":[{"description":"Item A","amount":120.00,"currency":"%s","category":"Food & Dining","subcategory":"Groceries","transaction_type":"expense","payment_method":"cash","movement_type":"transaction","counterparty":"","expense_date":"%s","merchant":""},{"description":"Item B","amount":80.00,"currency":"%s","category":"Food & Dining","subcategory":"Groceries","transaction_type":"expense","payment_method":"cash","movement_type":"transaction","counterparty":"","expense_date":"%s","merchant":""}],"not_expense_reply":""}
Single item or grand total: {"is_expense":true,"amount":200.00,"currency":"%s","description":"Grocery shopping","merchant":"Local Shop","category":"Food & Dining","subcategory":"Groceries","transaction_type":"expense","payment_method":"cash","movement_type":"transaction","counterparty":"","expense_date":"%s","not_expense_reply":"","items":[]}
Unreadable receipt: {"is_expense":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"Explain the problem","items":[]}

Rules:
- Use items[] ONLY for 2+ separate line items with individual prices; each item needs: description, amount, currency, category, subcategory, transaction_type, payment_method, expense_date, merchant.
- currency: default %s. Use ISO 4217 codes.
- expense_date: YYYY-MM-DD, default today if not on receipt.
- payment_method: "cash", "bank", "mobile_wallet", "debit_card", "credit_card", "card_unknown", or "other". Receipt says only "card" → "card_unknown".
- description: short English phrase.
- amount: positive number.`,
		today, cats,
		today, defaultCurrency, today, defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency)
}

// buildChatSystemPrompt returns the system prompt for multi-turn chat mode.
// Unlike buildSystemPrompt it allows the model to request more context via needs_context=true.
func buildChatSystemPrompt(categories []string, defaultCurrency string) string {
	today := time.Now().Format("2006-01-02")
	thisMonthStart := time.Now().Format("2006-01") + "-01"
	cats := "none yet"
	if len(categories) > 0 {
		cats = strings.Join(categories, ", ")
	}
	return fmt.Sprintf(`You are a personal finance assistant. Respond ONLY with a valid JSON object — no markdown, no extra text.
Today is %s. Categories: %s.

Transaction: {"is_expense":true,"needs_context":false,"amount":200.00,"currency":"%s","description":"English description","merchant":"","category":"Food","subcategory":"Lunch","transaction_type":"expense","payment_method":"cash","movement_type":"transaction","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"","query_category":"","query_from":"","query_to":"","delete_keywords":""}
Needs info: {"is_expense":false,"needs_context":true,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"Ask one short question in same language","query_type":"","query_category":"","query_from":"","query_to":"","delete_keywords":""}
CC bill/loan: {"is_expense":false,"needs_context":false,"amount":5000.00,"currency":"%s","description":"Loan repayment","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"loan_repayment","counterparty":"Rafi","expense_date":"%s","not_expense_reply":"","query_type":"","query_category":"","query_from":"","query_to":"","delete_keywords":""}
Query: {"is_expense":false,"needs_context":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"spending_sum","query_category":"Food","query_from":"%s","query_to":"%s","delete_keywords":""}
Delete: {"is_expense":false,"needs_context":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"","query_type":"delete_search","query_category":"","query_from":"%s","query_to":"%s","delete_keywords":"lunch"}
Chat/greeting: {"is_expense":false,"needs_context":false,"amount":0,"currency":"%s","description":"","merchant":"","category":"","subcategory":"","transaction_type":"expense","payment_method":"cash","movement_type":"","counterparty":"","expense_date":"%s","not_expense_reply":"Reply in same language","query_type":"","query_category":"","query_from":"","query_to":"","delete_keywords":""}

Rules:
- Use conversation history to fill in missing details across turns.
- transaction_type: "expense" for outgoing; "income" for salary, freelance, gifts, refunds.
- currency: default "%s". Detect local currency symbols/words and map to ISO codes.
- payment_method: "cash", "bank", "mobile_wallet", "debit_card", "credit_card", "card_unknown", or "other". Generic "card" → "card_unknown".
- movement_type: "transaction" for ordinary expense/income; "credit_card_payment" for CC bill; "loan_received" for borrowing; "loan_repayment" for repaying; "ambiguous_card" for generic card.
- CC bill payment (bill/due/minimum): is_expense=false, movement_type="credit_card_payment", payment_method="bank" or "cash".
- Card payment with no purchase description → needs_context=true, ask what was bought.
- CC bill or loan with no amount → needs_context=true, ask for amount.
- needs_context=true ONLY when transaction is clear but amount OR subject is missing.
- expense_date: today (%s) if not mentioned; YYYY-MM-DD.
- description: always English. not_expense_reply: same language as user.
- delete_keywords: keyword(s) to search. query_from/query_to narrow by date.`,
		today, cats,
		defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency, today,
		defaultCurrency, today, thisMonthStart, today,
		defaultCurrency, today, thisMonthStart, today,
		defaultCurrency, today,
		defaultCurrency,
		today)
}

func stripJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
