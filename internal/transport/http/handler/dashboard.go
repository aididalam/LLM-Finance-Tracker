package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
)

type DashboardHandler struct {
	expenseSvc *service.ExpenseService
	walletSvc  *service.WalletService
}

func NewDashboardHandler(expenseSvc *service.ExpenseService, walletSvc ...*service.WalletService) *DashboardHandler {
	h := &DashboardHandler{expenseSvc: expenseSvc}
	if len(walletSvc) > 0 {
		h.walletSvc = walletSvc[0]
	}
	return h
}

func (h *DashboardHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	expSvc := h.expenseSvc.ForUser(userID)

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = v
		}
	}

	expenseTotal, err := expSvc.MonthlyTotal(year, month)
	if err != nil {
		response.ResError(w, "failed to fetch overview", http.StatusInternalServerError)
		return
	}

	var incomeTotal float64
	if h.walletSvc != nil {
		incomeTotal, err = h.walletSvc.ForUser(userID).MonthlyIncomeTotal(year, month)
		if err != nil {
			response.ResError(w, "failed to fetch income total", http.StatusInternalServerError)
			return
		}
	}

	prev := time.Date(year, time.Month(month)-1, 1, 0, 0, 0, 0, time.Local)
	lastExpenseTotal, err := expSvc.MonthlyTotal(prev.Year(), int(prev.Month()))
	if err != nil {
		response.ResError(w, "failed to fetch overview", http.StatusInternalServerError)
		return
	}

	var lastIncomeTotal float64
	if h.walletSvc != nil {
		lastIncomeTotal, err = h.walletSvc.ForUser(userID).MonthlyIncomeTotal(prev.Year(), int(prev.Month()))
		if err != nil {
			response.ResError(w, "failed to fetch last income total", http.StatusInternalServerError)
			return
		}
	}

	expenseSummary, err := expSvc.MonthlySummary(year, month)
	if err != nil {
		response.ResError(w, "failed to fetch overview", http.StatusInternalServerError)
		return
	}

	net := incomeTotal - expenseTotal
	lastNet := lastIncomeTotal - lastExpenseTotal

	response.ResSuccess(w, map[string]any{
		"this_month":        expenseTotal,
		"last_month":        lastExpenseTotal,
		"income_this_month": incomeTotal,
		"income_last_month": lastIncomeTotal,
		"net_this_month":    net,
		"net_last_month":    lastNet,
		"categories":        expenseSummary.Categories,
	})
}

func (h *DashboardHandler) Trend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	expSvc := h.expenseSvc.ForUser(userID)

	now := time.Now()
	type point struct {
		Month   string  `json:"month"`
		Expense float64 `json:"expense"`
		Income  float64 `json:"income"`
		Net     float64 `json:"net"`
		Total   float64 `json:"total"`
	}
	points := make([]point, 12)
	for i := 11; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		expenseTotal, err := expSvc.MonthlyTotal(t.Year(), int(t.Month()))
		if err != nil {
			response.ResError(w, "failed to fetch trend", http.StatusInternalServerError)
			return
		}
		var incomeTotal float64
		if h.walletSvc != nil {
			incomeTotal, err = h.walletSvc.ForUser(userID).MonthlyIncomeTotal(t.Year(), int(t.Month()))
			if err != nil {
				response.ResError(w, "failed to fetch trend", http.StatusInternalServerError)
				return
			}
		}
		points[11-i] = point{
			Month:   t.Format("Jan 2006"),
			Expense: expenseTotal,
			Income:  incomeTotal,
			Net:     incomeTotal - expenseTotal,
			Total:   expenseTotal,
		}
	}
	response.ResSuccess(w, points)
}

