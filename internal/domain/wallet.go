package domain

import "time"

// Wallet is the top-level account (cash, bank, or MFS).
type Wallet struct {
	ID          string    `db:"id"           json:"id"`
	UserID      string    `db:"user_id"      json:"user_id"`
	Name        string    `db:"name"         json:"name"`
	AccountType string    `db:"account_type" json:"account_type"` // cash | bank | mfs
	IsActive    bool      `db:"is_active"    json:"is_active"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

// WalletWithBalance is a Wallet enriched with its computed balance.
type WalletWithBalance struct {
	Wallet
	Balance float64 `db:"balance" json:"balance"`
}

// WalletBank holds bank account details for a bank-type wallet.
type WalletBank struct {
	ID            string    `db:"id"              json:"id"`
	WalletID      string    `db:"wallet_id"       json:"wallet_id"`
	IsActive      bool      `db:"is_active"       json:"is_active"`
	BankName      string    `db:"bank_name"       json:"bank_name"`
	Branch        *string   `db:"branch"          json:"branch"`
	AccountNumber string    `db:"account_numeber" json:"account_number"` // column name kept per schema
	CreatedAt     time.Time `db:"created_at"      json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"      json:"updated_at"`
}

// WalletDebitCard is a debit card linked to a bank wallet.
type WalletDebitCard struct {
	ID           string    `db:"id"             json:"id"`
	WalletBankID string    `db:"wallet_bank_id" json:"wallet_bank_id"`
	WalletID     string    `db:"wallet_id"      json:"wallet_id"`
	IsActive     bool      `db:"is_active"      json:"is_active"`
	Last4Digit   string    `db:"last_4_digit"   json:"last_4_digit"`
	CreatedAt    time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updated_at"`
}

// WalletExpenseTransaction links an expense to the wallet it was paid from.
type WalletExpenseTransaction struct {
	ID                    string     `db:"id"            json:"id"`
	WalletID              string     `db:"wallet_id"     json:"wallet_id"`
	ExpenseID             string     `db:"expense_id"    json:"expense_id"`
	WalletBankDebitCardID *string    `db:"wallet_bank_debit_card_id" json:"wallet_bank_debit_card_id"`
	IsDeleted             bool       `db:"is_deleted"    json:"is_deleted"`
	DeletedAt             *time.Time `db:"deleted_at"    json:"deleted_at"`
	CreatedAt             time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"    json:"updated_at"`
}

// WalletInternalTransaction represents initial balance, income, or internal transfers.
type WalletInternalTransaction struct {
	ID         string     `db:"id"          json:"id"`
	WalletID   string     `db:"wallet_id"   json:"wallet_id"`
	SourceType string     `db:"source_type" json:"source_type"` // initial | income | internal
	ToWalletID *string    `db:"to_wallet_id" json:"to_wallet_id"`
	DeletedAt  *time.Time `db:"deleted_at"  json:"deleted_at"`
	Amount     float64    `db:"amount"      json:"amount"`
	Fees       float64    `db:"fees"        json:"fees"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
}

// WalletRepository is the persistence interface for wallets and sub-tables.
type WalletRepository interface {
	// Wallet CRUD
	List(userID string) ([]*Wallet, error)
	ListWithBalances(userID string) ([]*WalletWithBalance, error)
	FindByID(userID, id string) (*Wallet, error)
	Create(w *Wallet) error
	Update(w *Wallet) error
	Deactivate(userID, id string) error

	// Debit cards (bank accounts only)
	GetBankDetails(walletID string) (*WalletBank, error)
	UpsertBankDetails(bank *WalletBank) error
	GetDebitCards(walletID string) ([]*WalletDebitCard, error)
	AddDebitCard(card *WalletDebitCard) error

	// Transactions
	CreateInternalTransaction(tx *WalletInternalTransaction) error
	SoftDeleteExpenseTransaction(expenseID string) error

	// Queries
	GetBalance(userID, walletID string) (float64, error)
	MonthlyIncomeTotal(userID string, year, month int) (float64, error)
}
