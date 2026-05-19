package domain

import "time"

type Expense struct {
	ID              string     `db:"id"               json:"id"`
	UserID          string     `db:"user_id"          json:"user_id"`
	CategoryID      string     `db:"category_id"      json:"category_id"`
	SubcategoryID   string     `db:"subcategory_id"   json:"subcategory_id"`
	Amount          float64    `db:"amount"           json:"amount"`
	Fees            float64    `db:"fees"             json:"fees"`
	Description     *string    `db:"description"      json:"description"`
	ExpenseDatetime time.Time  `db:"expense_datetime" json:"expense_datetime"`
	ReceiptType     string     `db:"receipt_type"     json:"receipt_type"` // text | pdf | image
	IsDeleted       bool       `db:"is_deleted"       json:"is_deleted"`
	DeletedAt       *time.Time `db:"deleted_at"       json:"deleted_at"`
	CreatedAt       time.Time  `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"       json:"updated_at"`
}

type ExpenseRepository interface {
	// Create saves an expense and its wallet link atomically.
	Create(e *Expense, walletID string, walletBankDebitCardID *string) error
	FindAll(userID string, filters ExpenseFilters) ([]*Expense, error)
	FindAllWithCategory(userID string, filters ExpenseFilters) ([]*ExpenseWithCategory, error)
	FindByID(userID, id string) (*Expense, error)
	Update(e *Expense) error
	// SoftDelete soft-deletes the expense and its wallet_expense_transctions row.
	SoftDelete(userID, id string) error
	MonthlySummary(userID string, year, month int) (*MonthlySummary, error)
	MonthlyTotal(userID string, year, month int) (float64, error)
	SpentByCategoryForMonth(userID string, year, month int) (map[string]float64, error)
}

type ExpenseFilters struct {
	CategoryID *string
	Search     *string
	From       *string
	To         *string
	Limit      int
	Offset     int
}

// ExpenseWithCategory is returned for list endpoints that need category display info.
type ExpenseWithCategory struct {
	Expense
	CategoryName    string  `db:"category_name"    json:"category_name"`
	CategoryIcon    string  `db:"category_icon"    json:"category_icon"`
	SubcategoryName string  `db:"subcategory_name" json:"subcategory_name"`
	WalletID        *string `db:"wallet_id"        json:"wallet_id"`
	WalletName      *string `db:"wallet_name"      json:"wallet_name"`
}

type MonthlySummary struct {
	Total      float64         `json:"total"`
	Categories []CategoryTotal `json:"categories"`
}

type CategoryTotal struct {
	CategoryName string  `db:"category_name" json:"category_name"`
	Icon         string  `db:"icon"          json:"icon"`
	Total        float64 `db:"total"         json:"total"`
	Count        int     `db:"count"         json:"count"`
}
