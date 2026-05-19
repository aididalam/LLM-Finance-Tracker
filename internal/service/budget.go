package service

import (
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/google/uuid"
)

type BudgetService struct {
	userID      string
	repo        domain.BudgetRepository
	expenseRepo domain.ExpenseRepository
}

func NewBudgetService(repo domain.BudgetRepository, expenseRepo domain.ExpenseRepository) *BudgetService {
	return &BudgetService{repo: repo, expenseRepo: expenseRepo, userID: domain.DefaultUserID}
}

func (s *BudgetService) ForUser(userID string) *BudgetService {
	scoped := *s
	scoped.userID = userID
	return &scoped
}

func (s *BudgetService) Upsert(categoryID *string, amount float64, year, month int, carryOver bool) (*domain.Budget, error) {
	existing, err := s.repo.FindByCategoryMonth(s.userID, categoryID, year, month)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		existing.Amount = amount
		existing.CarryOver = carryOver
		return existing, s.repo.Upsert(existing)
	}

	b := &domain.Budget{
		ID:         uuid.New().String(),
		UserID:     s.userID,
		CategoryID: categoryID,
		Amount:     amount,
		Month:      month,
		Year:       year,
		CarryOver:  carryOver,
	}
	return b, s.repo.Upsert(b)
}

// StatusForMonth returns all budgets for the given month enriched with spending and carry-over.
func (s *BudgetService) StatusForMonth(year, month int) ([]*domain.BudgetStatus, error) {
	budgets, err := s.repo.FindByMonth(s.userID, year, month)
	if err != nil {
		return nil, err
	}

	summary, err := s.expenseRepo.MonthlySummary(s.userID, year, month)
	if err != nil {
		return nil, err
	}

	// Build a map of category spent amounts
	spentByCategory := map[string]float64{}
	for _, ct := range summary.Categories {
		// ct.CategoryName is used for display, but we need category_id matching
		// We'll match by position - use the full summary query which uses category_id
		_ = ct
	}
	// Refetch total per category_id
	spentByCatID, err := s.expenseRepo.SpentByCategoryForMonth(s.userID, year, month)
	if err != nil {
		return nil, err
	}
	for catID, total := range spentByCatID {
		spentByCategory[catID] = total
	}

	var statuses []*domain.BudgetStatus
	for _, b := range budgets {
		var spent float64
		if b.CategoryID != nil {
			spent = spentByCategory[*b.CategoryID]
		} else {
			// Overall budget — sum everything
			for _, v := range spentByCategory {
				spent += v
			}
		}

		effective := b.Amount
		if b.CarryOver {
			// Add leftover from previous month
			prevMonth := time.Date(year, time.Month(month)-1, 1, 0, 0, 0, 0, time.Local)
			prevBudget, _ := s.repo.FindByCategoryMonth(s.userID, b.CategoryID, prevMonth.Year(), int(prevMonth.Month()))
			if prevBudget != nil {
				var prevSpent float64
				if b.CategoryID != nil {
					prevSpentMap, _ := s.expenseRepo.SpentByCategoryForMonth(s.userID, prevMonth.Year(), int(prevMonth.Month()))
					prevSpent = prevSpentMap[*b.CategoryID]
				}
				leftover := prevBudget.Amount - prevSpent
				if leftover > 0 {
					effective += leftover
				}
			}
		}

		pct := 0.0
		if effective > 0 {
			pct = (spent / effective) * 100
		}

		statuses = append(statuses, &domain.BudgetStatus{
			Budget:    *b,
			Spent:     spent,
			Effective: effective,
			Pct:       pct,
		})
	}
	return statuses, nil
}
