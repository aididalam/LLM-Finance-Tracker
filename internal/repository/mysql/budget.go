package mysql

import (
	"database/sql"
	"errors"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/jmoiron/sqlx"
)

type budgetRepository struct {
	db *sqlx.DB
}

func NewBudgetRepository(db *sqlx.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Upsert(b *domain.Budget) error {
	_, err := r.db.NamedExec(`
		INSERT INTO budgets (budget_id, user_id, category_id, amount, month, year, carry_over)
		VALUES (:budget_id, :user_id, :category_id, :amount, :month, :year, :carry_over)
		ON DUPLICATE KEY UPDATE amount = VALUES(amount), carry_over = VALUES(carry_over)
	`, b)
	return err
}

func (r *budgetRepository) FindByMonth(userID string, year, month int) ([]*domain.Budget, error) {
	var budgets []*domain.Budget
	err := r.db.Select(&budgets,
		"SELECT * FROM budgets WHERE user_id = ? AND year = ? AND month = ?", userID, year, month)
	return budgets, err
}

func (r *budgetRepository) FindByCategoryMonth(userID string, categoryID *string, year, month int) (*domain.Budget, error) {
	var b domain.Budget
	var err error
	if categoryID == nil {
		err = r.db.Get(&b,
			"SELECT * FROM budgets WHERE user_id = ? AND category_id IS NULL AND year = ? AND month = ? LIMIT 1",
			userID, year, month)
	} else {
		err = r.db.Get(&b,
			"SELECT * FROM budgets WHERE user_id = ? AND category_id = ? AND year = ? AND month = ? LIMIT 1",
			userID, *categoryID, year, month)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}
