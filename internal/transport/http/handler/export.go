package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
)

type ExportHandler struct {
	expenseSvc *service.ExpenseService
}

func NewExportHandler(expenseSvc *service.ExpenseService) *ExportHandler {
	return &ExportHandler{expenseSvc: expenseSvc}
}

func (h *ExportHandler) CSV(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.expenseSvc.ForUser(userID)

	q := r.URL.Query()
	filters := domain.ExpenseFilters{Limit: 10000}

	if v := q.Get("from"); v != "" {
		filters.From = &v
	}
	if v := q.Get("to"); v != "" {
		filters.To = &v
	}
	if v := q.Get("category_id"); v != "" {
		filters.CategoryID = &v
	}

	// Default to current month if no range given
	if filters.From == nil && filters.To == nil {
		now := time.Now()
		from := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
		to := now.Format("2006-01-02")
		filters.From = &from
		filters.To = &to
	}

	expenses, err := svc.FindAllWithCategory(filters)
	if err != nil {
		response.ResError(w, "export failed", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("expenses_%s.csv", time.Now().Format("2006-01"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Date", "Description", "Category", "Subcategory", "Amount", "Fees", "Receipt Type", "Wallet"})

	for _, e := range expenses {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		walletName := ""
		if e.WalletName != nil {
			walletName = *e.WalletName
		}
		_ = cw.Write([]string{
			e.ExpenseDatetime.Format("2006-01-02"),
			desc,
			e.CategoryName,
			e.SubcategoryName,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			strconv.FormatFloat(e.Fees, 'f', 2, 64),
			e.ReceiptType,
			walletName,
		})
	}
	cw.Flush()
}
