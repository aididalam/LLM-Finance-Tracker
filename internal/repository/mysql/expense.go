package mysql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type expenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) domain.ExpenseRepository {
	return &expenseRepository{db: db}
}

// Create saves an expense and its wallet_expense_transctions row in a single DB transaction.
func (r *expenseRepository) Create(e *domain.Expense, walletID string, walletBankDebitCardID *string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.NamedExec(`
		INSERT INTO expense
			(id, user_id, category_id, subcategory_id, amount, fees, description,
			 expense_datetime, receipt_type)
		VALUES
			(:id, :user_id, :category_id, :subcategory_id, :amount, :fees, :description,
			 :expense_datetime, :receipt_type)
	`, e)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO wallet_expense_transctions (id, wallet_id, expense_id, wallet_bank_debit_card_id)
		VALUES (?, ?, ?, ?)
	`, uuid.New().String(), walletID, e.ID, walletBankDebitCardID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *expenseRepository) FindAll(userID string, f domain.ExpenseFilters) ([]*domain.Expense, error) {
	q := "SELECT * FROM expense WHERE user_id = ? AND is_deleted = FALSE"
	args := []any{userID}

	q, args = applyExpenseFilters(q, args, f)
	q += " ORDER BY expense_datetime DESC, created_at DESC"
	q, args = applyPaging(q, args, f.Limit, f.Offset)

	var expenses []*domain.Expense
	err := r.db.Select(&expenses, q, args...)
	return expenses, err
}

func (r *expenseRepository) FindAllWithCategory(userID string, f domain.ExpenseFilters) ([]*domain.ExpenseWithCategory, error) {
	q := `
		SELECT
			e.*,
			COALESCE(c.name,'')  AS category_name,
			COALESCE(c.icon,'')  AS category_icon,
			COALESCE(sc.name,'') AS subcategory_name,
			wet.wallet_id        AS wallet_id,
			w.name               AS wallet_name
		FROM expense e
		LEFT JOIN categories     c  ON c.category_id    = e.category_id
		LEFT JOIN subcategories  sc ON sc.subcategory_id = e.subcategory_id
		LEFT JOIN wallet_expense_transctions wet
			ON wet.expense_id = e.id AND wet.is_deleted = FALSE
		LEFT JOIN wallet w ON w.id = wet.wallet_id
		WHERE e.user_id = ? AND e.is_deleted = FALSE`
	args := []any{userID}

	q, args = applyExpenseFilters(q, args, f)
	q += " ORDER BY e.expense_datetime DESC, e.created_at DESC"

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	q, args = applyPaging(q, args, limit, f.Offset)

	var expenses []*domain.ExpenseWithCategory
	err := r.db.Select(&expenses, q, args...)
	return expenses, err
}

func (r *expenseRepository) FindByID(userID, id string) (*domain.Expense, error) {
	var e domain.Expense
	err := r.db.Get(&e, `
		SELECT * FROM expense WHERE user_id = ? AND id = ? AND is_deleted = FALSE LIMIT 1
	`, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &e, err
}

func (r *expenseRepository) Update(e *domain.Expense) error {
	_, err := r.db.NamedExec(`
		UPDATE expense
		SET category_id      = :category_id,
		    subcategory_id   = :subcategory_id,
		    amount           = :amount,
		    fees             = :fees,
		    description      = :description,
		    expense_datetime = :expense_datetime,
		    receipt_type     = :receipt_type
		WHERE user_id = :user_id AND id = :id AND is_deleted = FALSE
	`, e)
	return err
}

func (r *expenseRepository) SoftDelete(userID, id string) error {
	now := time.Now()
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		UPDATE expense SET is_deleted = TRUE, deleted_at = ?
		WHERE user_id = ? AND id = ?
	`, now, userID, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE wallet_expense_transctions
		SET is_deleted = TRUE, deleted_at = ?
		WHERE expense_id = ?
	`, now, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *expenseRepository) SpentByCategoryForMonth(userID string, year, month int) (map[string]float64, error) {
	rows, err := r.db.Queryx(`
		SELECT category_id AS cat_id, SUM(amount) AS total
		FROM expense
		WHERE YEAR(expense_datetime) = ? AND MONTH(expense_datetime) = ?
		  AND user_id = ?
		  AND is_deleted = FALSE
		GROUP BY category_id
	`, year, month, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]float64{}
	for rows.Next() {
		var catID string
		var total float64
		if err := rows.Scan(&catID, &total); err != nil {
			return nil, err
		}
		result[catID] = total
	}
	return result, nil
}

func (r *expenseRepository) MonthlySummary(userID string, year, month int) (*domain.MonthlySummary, error) {
	total, err := r.MonthlyTotal(userID, year, month)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Queryx(`
		SELECT COALESCE(c.name,'Uncategorized') AS category_name,
		       COALESCE(c.icon,'📦')            AS icon,
		       SUM(e.amount)                    AS total,
		       COUNT(*)                         AS count
		FROM expense e
		LEFT JOIN categories c ON c.category_id = e.category_id
		WHERE YEAR(e.expense_datetime) = ? AND MONTH(e.expense_datetime) = ?
		  AND e.user_id = ?
		  AND e.is_deleted = FALSE
		GROUP BY e.category_id, c.name, c.icon
		ORDER BY total DESC
	`, year, month, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []domain.CategoryTotal
	for rows.Next() {
		var ct domain.CategoryTotal
		if err := rows.StructScan(&ct); err != nil {
			return nil, err
		}
		cats = append(cats, ct)
	}

	return &domain.MonthlySummary{
		Total:      total,
		Categories: cats,
	}, nil
}

func (r *expenseRepository) MonthlyTotal(userID string, year, month int) (float64, error) {
	var total sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT SUM(amount) FROM expense
		WHERE YEAR(expense_datetime) = ? AND MONTH(expense_datetime) = ?
		  AND user_id = ?
		  AND is_deleted = FALSE
	`, year, month, userID).Scan(&total)
	return total.Float64, err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func applyExpenseFilters(q string, args []any, f domain.ExpenseFilters) (string, []any) {
	if f.CategoryID != nil {
		q += " AND e.category_id = ?"
		args = append(args, *f.CategoryID)
	}
	if f.Search != nil && *f.Search != "" {
		q += " AND (e.description LIKE ? OR sc.name LIKE ?)"
		s := "%" + *f.Search + "%"
		args = append(args, s, s)
	}
	if f.From != nil {
		q += " AND e.expense_datetime >= ?"
		args = append(args, *f.From)
	}
	if f.To != nil {
		q += " AND e.expense_datetime <= ?"
		args = append(args, *f.To)
	}
	return q, args
}

func applyPaging(q string, args []any, limit, offset int) (string, []any) {
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		q += " OFFSET ?"
		args = append(args, offset)
	}
	return q, args
}
