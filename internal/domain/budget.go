package domain

import "time"

type Budget struct {
	ID         string    `db:"budget_id"   json:"budget_id"`
	UserID     string    `db:"user_id"     json:"user_id"`
	CategoryID *string   `db:"category_id" json:"category_id"`
	Amount     float64   `db:"amount"      json:"amount"`
	Month      int       `db:"month"        json:"month"`
	Year       int       `db:"year"         json:"year"`
	CarryOver  bool      `db:"carry_over"  json:"carry_over"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"  json:"updated_at"`
}

// BudgetStatus combines a budget with current spending for alert/display purposes.
type BudgetStatus struct {
	Budget
	CategoryName string  `db:"category_name" json:"category_name"`
	CategoryIcon string  `db:"category_icon" json:"category_icon"`
	Spent        float64 `json:"spent"`
	Effective    float64 `json:"effective"` // budget + carry-over from last month
	Pct          float64 `json:"pct"`
}

type BudgetRepository interface {
	Upsert(b *Budget) error
	FindByMonth(userID string, year, month int) ([]*Budget, error)
	FindByCategoryMonth(userID string, categoryID *string, year, month int) (*Budget, error)
}
