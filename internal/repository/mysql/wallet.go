package mysql

import (
	"database/sql"
	"errors"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/jmoiron/sqlx"
)

type walletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

// ── Wallet CRUD ──────────────────────────────────────────────────────────────

func (r *walletRepository) List(userID string) ([]*domain.Wallet, error) {
	var wallets []*domain.Wallet
	err := r.db.Select(&wallets, `
		SELECT * FROM wallet
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY FIELD(account_type,'cash','bank','mfs'), name ASC
	`, userID)
	return wallets, err
}

func (r *walletRepository) ListWithBalances(userID string) ([]*domain.WalletWithBalance, error) {
	var rows []*domain.WalletWithBalance
	err := r.db.Select(&rows, `
		SELECT
			w.id, w.user_id, w.name, w.account_type, w.is_active,
			w.created_at, w.updated_at,
			COALESCE(
				(SELECT SUM(wit.amount - wit.fees)
				 FROM wallet_internal_transctions wit
				 WHERE wit.wallet_id = w.id AND wit.source_type IN ('initial','income') AND wit.deleted_at IS NULL),
			0)
			+
			COALESCE(
				(SELECT SUM(wit.amount)
				 FROM wallet_internal_transctions wit
				 WHERE wit.to_wallet_id = w.id AND wit.source_type = 'internal' AND wit.deleted_at IS NULL),
			0)
			-
			COALESCE(
				(SELECT SUM(wit.amount + wit.fees)
				 FROM wallet_internal_transctions wit
				 WHERE wit.wallet_id = w.id AND wit.source_type = 'internal' AND wit.to_wallet_id IS NOT NULL AND wit.deleted_at IS NULL),
			0)
			-
			COALESCE(
				(SELECT SUM(e.amount + e.fees)
				 FROM wallet_expense_transctions wet
				 INNER JOIN expense e ON e.id = wet.expense_id
				 WHERE wet.wallet_id = w.id
				   AND wet.is_deleted = FALSE AND wet.deleted_at IS NULL
				   AND e.is_deleted = FALSE AND e.deleted_at IS NULL),
			0) AS balance
		FROM wallet w
		WHERE w.user_id = ? AND w.is_active = TRUE
		ORDER BY FIELD(w.account_type,'cash','bank','mfs'), w.name ASC
	`, userID)
	return rows, err
}

func (r *walletRepository) FindByID(userID, id string) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.db.Get(&w, `SELECT * FROM wallet WHERE user_id = ? AND id = ? LIMIT 1`, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

func (r *walletRepository) Create(w *domain.Wallet) error {
	_, err := r.db.NamedExec(`
		INSERT INTO wallet (id, user_id, name, account_type, is_active)
		VALUES (:id, :user_id, :name, :account_type, :is_active)
	`, w)
	return err
}

func (r *walletRepository) Update(w *domain.Wallet) error {
	_, err := r.db.NamedExec(`
		UPDATE wallet SET name = :name, account_type = :account_type
		WHERE user_id = :user_id AND id = :id
	`, w)
	return err
}

func (r *walletRepository) Deactivate(userID, id string) error {
	_, err := r.db.Exec(`UPDATE wallet SET is_active = FALSE WHERE user_id = ? AND id = ?`, userID, id)
	return err
}

// ── Debit cards ──────────────────────────────────────────────────────────────

func (r *walletRepository) GetBankDetails(walletID string) (*domain.WalletBank, error) {
	var b domain.WalletBank
	err := r.db.Get(&b, `SELECT * FROM wallet_bank WHERE wallet_id = ? LIMIT 1`, walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *walletRepository) UpsertBankDetails(bank *domain.WalletBank) error {
	_, err := r.db.NamedExec(`
		INSERT INTO wallet_bank (id, wallet_id, is_active, bank_name, branch, account_numeber)
		VALUES (:id, :wallet_id, :is_active, :bank_name, :branch, :account_numeber)
		ON DUPLICATE KEY UPDATE
			bank_name = VALUES(bank_name),
			branch = VALUES(branch),
			account_numeber = VALUES(account_numeber),
			is_active = VALUES(is_active)
	`, bank)
	return err
}

func (r *walletRepository) GetDebitCards(walletID string) ([]*domain.WalletDebitCard, error) {
	var cards []*domain.WalletDebitCard
	err := r.db.Select(&cards, `SELECT * FROM wallet_bank_debit_card WHERE wallet_id = ? ORDER BY created_at ASC`, walletID)
	if cards == nil {
		cards = []*domain.WalletDebitCard{}
	}
	return cards, err
}

func (r *walletRepository) AddDebitCard(card *domain.WalletDebitCard) error {
	_, err := r.db.NamedExec(`
		INSERT INTO wallet_bank_debit_card (id, wallet_bank_id, wallet_id, is_active, last_4_digit)
		VALUES (:id, :wallet_bank_id, :wallet_id, :is_active, :last_4_digit)
	`, card)
	return err
}

// ── Transactions ─────────────────────────────────────────────────────────────

func (r *walletRepository) CreateInternalTransaction(tx *domain.WalletInternalTransaction) error {
	_, err := r.db.NamedExec(`
		INSERT INTO wallet_internal_transctions
			(id, wallet_id, source_type, to_wallet_id, amount, fees)
		VALUES
			(:id, :wallet_id, :source_type, :to_wallet_id, :amount, :fees)
	`, tx)
	return err
}

func (r *walletRepository) SoftDeleteExpenseTransaction(expenseID string) error {
	_, err := r.db.Exec(`
		UPDATE wallet_expense_transctions
		SET is_deleted = TRUE, deleted_at = NOW()
		WHERE expense_id = ?
	`, expenseID)
	return err
}

// ── Queries ───────────────────────────────────────────────────────────────────

func (r *walletRepository) GetBalance(userID, walletID string) (float64, error) {
	var balance sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT
			COALESCE(
				(SELECT SUM(wit.amount - wit.fees)
				 FROM wallet_internal_transctions wit
				 WHERE wit.wallet_id = ? AND wit.source_type IN ('initial','income') AND wit.deleted_at IS NULL),
			0)
			+
			COALESCE(
				(SELECT SUM(wit.amount)
				 FROM wallet_internal_transctions wit
				 WHERE wit.to_wallet_id = ? AND wit.source_type = 'internal' AND wit.deleted_at IS NULL),
			0)
			-
			COALESCE(
				(SELECT SUM(wit.amount + wit.fees)
				 FROM wallet_internal_transctions wit
				 WHERE wit.wallet_id = ? AND wit.source_type = 'internal' AND wit.to_wallet_id IS NOT NULL AND wit.deleted_at IS NULL),
			0)
			-
			COALESCE(
				(SELECT SUM(e.amount + e.fees)
				 FROM wallet_expense_transctions wet
				 INNER JOIN expense e ON e.id = wet.expense_id
				 WHERE wet.wallet_id = ?
				   AND wet.is_deleted = FALSE AND wet.deleted_at IS NULL
				   AND e.is_deleted = FALSE AND e.deleted_at IS NULL),
			0) AS balance
		FROM wallet w
		WHERE w.user_id = ? AND w.id = ?
	`, walletID, walletID, walletID, walletID, userID, walletID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance.Float64, nil
}

func (r *walletRepository) MonthlyIncomeTotal(userID string, year, month int) (float64, error) {
	var total sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT SUM(wit.amount - wit.fees)
		FROM wallet_internal_transctions wit
		INNER JOIN wallet w ON w.id = wit.wallet_id
		WHERE w.user_id = ?
		  AND wit.source_type = 'income'
		  AND wit.deleted_at IS NULL
		  AND YEAR(wit.created_at) = ?
		  AND MONTH(wit.created_at) = ?
	`, userID, year, month).Scan(&total)
	return total.Float64, err
}
