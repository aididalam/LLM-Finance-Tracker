package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
)

type ExpenseHandler struct {
	svc            *service.ExpenseService
	subcategorySvc *service.SubcategoryService
	llmProvider    llm.Provider
	categorySvc    *service.CategoryService
	llmUsage       *mysql.LLMUsageRepository
}

func NewExpenseHandler(svc *service.ExpenseService, subcategorySvc *service.SubcategoryService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc, subcategorySvc: subcategorySvc}
}

// WithLLM wires an LLM provider and category service for AI-powered parsing.
func (h *ExpenseHandler) WithLLM(p llm.Provider, categorySvc *service.CategoryService, llmUsage ...*mysql.LLMUsageRepository) *ExpenseHandler {
	h.llmProvider = p
	h.categorySvc = categorySvc
	if len(llmUsage) > 0 {
		h.llmUsage = llmUsage[0]
	}
	return h
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	var body struct {
		Description           string  `json:"description"`
		Amount                float64 `json:"amount"`
		Fees                  float64 `json:"fees"`
		CategoryID            string  `json:"category_id"`
		SubcategoryID         string  `json:"subcategory_id"`
		WalletID              string  `json:"wallet_id"`
		WalletBankDebitCardID *string `json:"wallet_bank_debit_card_id"`
		ExpenseDate           string  `json:"expense_date"`
		ReceiptType           string  `json:"receipt_type"`
		Subcategory           string  `json:"subcategory"` // name-based fallback
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body")
		return
	}
	if body.Amount <= 0 {
		response.ResError(w, "amount must be greater than zero")
		return
	}
	if body.CategoryID == "" {
		response.ResError(w, "category_id is required")
		return
	}
	if body.WalletID == "" {
		response.ResError(w, "wallet_id is required")
		return
	}

	date := time.Now()
	if body.ExpenseDate != "" {
		if d, err := time.Parse("2006-01-02", body.ExpenseDate); err == nil {
			date = d
		}
	}

	// Resolve subcategory by name when ID is not provided.
	subcategoryID := body.SubcategoryID
	subcatSvc := h.subcategorySvc
	if subcatSvc != nil {
		subcatSvc = subcatSvc.ForUser(userID)
	}
	if subcategoryID == "" && body.Subcategory != "" && subcatSvc != nil {
		catID := body.CategoryID
		subcat, err := subcatSvc.FindOrCreate(&catID, body.Subcategory)
		if err == nil && subcat != nil {
			subcategoryID = subcat.ID
		}
	}

	rt := "text"
	if body.ReceiptType == "image" || body.ReceiptType == "pdf" {
		rt = body.ReceiptType
	}

	e, err := svc.CreateManual(service.CreateManualParams{
		Description:           body.Description,
		Amount:                body.Amount,
		Fees:                  body.Fees,
		Date:                  date,
		CategoryID:            body.CategoryID,
		SubcategoryID:         subcategoryID,
		WalletID:              body.WalletID,
		WalletBankDebitCardID: body.WalletBankDebitCardID,
		ReceiptType:           rt,
	})
	if err != nil {
		response.ResError(w, "failed to create expense", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, e, http.StatusCreated)
}

// LLMParse parses a natural-language description using the LLM provider.
// Nothing is saved to the DB.
func (h *ExpenseHandler) LLMParse(w http.ResponseWriter, r *http.Request) {
	if h.llmProvider == nil || h.categorySvc == nil {
		response.ResError(w, "LLM not configured", http.StatusServiceUnavailable)
		return
	}
	userID := middleware.CurrentUserID(r.Context())
	categorySvc := h.categorySvc.ForUser(userID)

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		response.ResError(w, "text is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	categories, err := categorySvc.Names()
	if err != nil {
		response.ResError(w, "failed to load categories", http.StatusInternalServerError)
		return
	}

	parsed, usage, err := h.llmProvider.ParseExpense(ctx, body.Text, categories, "")
	if usage != nil && h.llmUsage != nil {
		_ = h.llmUsage.ForUser(userID).Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "web_llm_parse")
	}
	if err != nil {
		response.ResError(w, "LLM parse failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	parsed.Normalize()

	if !parsed.IsExpense {
		reply := parsed.NotExpenseReply
		if reply == "" {
			reply = "That doesn't look like a transaction. Try describing it differently."
		}
		response.ResSuccess(w, map[string]interface{}{
			"is_expense":       false,
			"is_income":        service.IsIncomeType(parsed.TransactionType),
			"reply":            reply,
			"transaction_type": parsed.TransactionType,
			"amount":           parsed.Amount,
			"expense_date":     parsed.ExpenseDate,
		})
		return
	}

	cat, err := categorySvc.FindOrCreate(parsed.Category)
	if err != nil {
		response.ResError(w, "failed to resolve category", http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"is_expense":       true,
		"is_income":        service.IsIncomeType(parsed.TransactionType),
		"transaction_type": parsed.TransactionType,
		"amount":           parsed.Amount,
		"description":      parsed.Description,
		"category":         parsed.Category,
		"category_id":      nil,
		"subcategory":      parsed.Subcategory,
		"expense_date":     parsed.ExpenseDate,
	}
	if cat != nil {
		result["category_id"] = cat.ID
		result["category"] = cat.Name
	}
	response.ResSuccess(w, result)
}

func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	q := r.URL.Query()
	filters := domain.ExpenseFilters{Limit: 20}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			filters.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filters.Offset = n
		}
	}
	if v := q.Get("category_id"); v != "" {
		filters.CategoryID = &v
	}
	if v := q.Get("q"); v != "" {
		filters.Search = &v
	}
	if v := q.Get("from"); v != "" {
		filters.From = &v
	}
	if v := q.Get("to"); v != "" {
		filters.To = &v
	}

	expenses, err := svc.FindAllWithCategory(filters)
	if err != nil {
		response.ResError(w, "failed to fetch expenses", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, expenses)
}

// BulkDelete soft-deletes multiple expenses by ID.
func (h *ExpenseHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		response.ResError(w, "ids array is required")
		return
	}
	deleted := 0
	for _, id := range body.IDs {
		if err := svc.Delete(id); err == nil {
			deleted++
		}
	}
	response.ResSuccess(w, map[string]int{"deleted": deleted})
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	id := chi.URLParam(r, "id")
	expense, err := svc.FindByID(id)
	if err != nil || expense == nil {
		response.ResError(w, "expense not found", http.StatusNotFound)
		return
	}
	if err := svc.Delete(id); err != nil {
		response.ResError(w, "failed to delete expense", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, "expense deleted")
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	id := chi.URLParam(r, "id")
	expense, err := svc.FindByID(id)
	if err != nil || expense == nil {
		response.ResError(w, "expense not found", http.StatusNotFound)
		return
	}

	var body struct {
		Amount        *float64 `json:"amount"`
		Fees          *float64 `json:"fees"`
		Description   *string  `json:"description"`
		CategoryID    *string  `json:"category_id"`
		SubcategoryID *string  `json:"subcategory_id"`
		Subcategory   *string  `json:"subcategory"` // name-based fallback
		ExpenseDate   *string  `json:"expense_date"`
		ReceiptType   *string  `json:"receipt_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body")
		return
	}

	subcatSvc := h.subcategorySvc
	if subcatSvc != nil {
		subcatSvc = subcatSvc.ForUser(userID)
	}

	if body.Amount != nil {
		expense.Amount = *body.Amount
	}
	if body.Fees != nil {
		expense.Fees = *body.Fees
	}
	if body.Description != nil {
		if *body.Description == "" {
			expense.Description = nil
		} else {
			expense.Description = body.Description
		}
	}
	if body.CategoryID != nil {
		expense.CategoryID = *body.CategoryID
	}
	if body.SubcategoryID != nil {
		expense.SubcategoryID = *body.SubcategoryID
	}
	if body.Subcategory != nil && *body.Subcategory != "" && body.SubcategoryID == nil && subcatSvc != nil {
		catID := expense.CategoryID
		subcat, err := subcatSvc.FindOrCreate(&catID, *body.Subcategory)
		if err == nil && subcat != nil {
			expense.SubcategoryID = subcat.ID
		}
	}
	if body.ExpenseDate != nil {
		if d, err := time.Parse("2006-01-02", *body.ExpenseDate); err == nil {
			expense.ExpenseDatetime = d
		}
	}
	if body.ReceiptType != nil {
		expense.ReceiptType = *body.ReceiptType
	}

	if err := svc.Update(expense); err != nil {
		response.ResError(w, "failed to update expense", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, expense)
}
