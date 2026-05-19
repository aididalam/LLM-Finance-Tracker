package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
)

type BudgetHandler struct {
	svc *service.BudgetService
}

func NewBudgetHandler(svc *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{svc: svc}
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	now := time.Now()
	year, month := now.Year(), int(now.Month())

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

	statuses, err := svc.StatusForMonth(year, month)
	if err != nil {
		response.ResError(w, "failed to fetch budgets", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, statuses)
}

func (h *BudgetHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	var body struct {
		CategoryID *string `json:"category_id"`
		Amount     float64 `json:"amount"`
		Month      int     `json:"month"`
		Year       int     `json:"year"`
		CarryOver  bool    `json:"carry_over"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Amount <= 0 {
		response.ResError(w, "amount must be positive", http.StatusBadRequest)
		return
	}

	now := time.Now()
	if body.Month == 0 {
		body.Month = int(now.Month())
	}
	if body.Year == 0 {
		body.Year = now.Year()
	}

	budget, err := svc.Upsert(body.CategoryID, body.Amount, body.Year, body.Month, body.CarryOver)
	if err != nil {
		response.ResError(w, "failed to save budget", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, budget)
}
