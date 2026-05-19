package service

import (
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/google/uuid"
)

// IsIncomeType returns true when the LLM-parsed transaction_type represents income.
func IsIncomeType(transactionType string) bool {
	switch strings.ToLower(strings.TrimSpace(transactionType)) {
	case "income", "earning", "earnings", "revenue", "salary", "credit", "received", "receive":
		return true
	default:
		return false
	}
}

type ExpenseService struct {
	userID         string
	repo           domain.ExpenseRepository
	subcategorySvc *SubcategoryService
	walletSvc      *WalletService
}

func NewExpenseService(repo domain.ExpenseRepository, subcategorySvc ...*SubcategoryService) *ExpenseService {
	svc := &ExpenseService{repo: repo, userID: domain.DefaultUserID}
	if len(subcategorySvc) > 0 {
		svc.subcategorySvc = subcategorySvc[0]
	}
	return svc
}

func (s *ExpenseService) ForUser(userID string) *ExpenseService {
	scoped := *s
	scoped.userID = userID
	if s.subcategorySvc != nil {
		scoped.subcategorySvc = s.subcategorySvc.ForUser(userID)
	}
	if s.walletSvc != nil {
		scoped.walletSvc = s.walletSvc.ForUser(userID)
	}
	return &scoped
}

func (s *ExpenseService) WithWallets(walletSvc *WalletService) *ExpenseService {
	s.walletSvc = walletSvc
	return s
}

// CreateFromParsed saves an LLM-parsed expense. Income transactions are
// delegated to WalletService.AddIncome rather than stored in the expense table.
// Returns (expense, isIncome, error).
func (s *ExpenseService) CreateFromParsed(
	parsed *llm.ParsedExpense,
	categoryID *string,
	walletID string,
	receiptType string,
	walletBankDebitCardID ...*string,
) (*domain.Expense, bool, error) {
	parsed.Normalize()

	if IsIncomeType(parsed.TransactionType) {
		if s.walletSvc != nil {
			if err := s.walletSvc.AddIncome(walletID, parsed.Amount, 0); err != nil {
				return nil, true, err
			}
		}
		return nil, true, nil
	}

	date, err := time.Parse("2006-01-02", parsed.ExpenseDate)
	if err != nil {
		date = time.Now()
	}

	var subcategoryID string
	if s.subcategorySvc != nil {
		subcat, err := s.subcategorySvc.FindOrCreate(categoryID, parsed.Subcategory)
		if err != nil {
			return nil, false, err
		}
		if subcat != nil {
			subcategoryID = subcat.ID
			parsed.Subcategory = subcat.Name
		}
	}

	var catID string
	if categoryID != nil {
		catID = *categoryID
	}

	var desc *string
	if d := strings.TrimSpace(parsed.Description); d != "" {
		desc = &d
	}

	rt := "text"
	if receiptType != "" {
		rt = receiptType
	}

	e := &domain.Expense{
		ID:              uuid.New().String(),
		UserID:          s.userID,
		CategoryID:      catID,
		SubcategoryID:   subcategoryID,
		Amount:          parsed.Amount,
		Fees:            0,
		Description:     desc,
		ExpenseDatetime: date,
		ReceiptType:     rt,
	}

	var debitCardID *string
	if len(walletBankDebitCardID) > 0 {
		debitCardID = walletBankDebitCardID[0]
	}

	if err := s.repo.Create(e, walletID, debitCardID); err != nil {
		return nil, false, err
	}
	return e, false, nil
}

// CreateManualParams holds parameters for a manually submitted expense.
type CreateManualParams struct {
	Description           string
	Amount                float64
	Fees                  float64
	Date                  time.Time
	CategoryID            string
	SubcategoryID         string
	WalletID              string
	WalletBankDebitCardID *string
	ReceiptType           string
}

// CreateManual creates an expense submitted from the web dashboard.
func (s *ExpenseService) CreateManual(p CreateManualParams) (*domain.Expense, error) {
	rt := "text"
	if p.ReceiptType != "" {
		rt = p.ReceiptType
	}

	var desc *string
	if d := strings.TrimSpace(p.Description); d != "" {
		desc = &d
	}

	e := &domain.Expense{
		ID:              uuid.New().String(),
		UserID:          s.userID,
		CategoryID:      p.CategoryID,
		SubcategoryID:   p.SubcategoryID,
		Amount:          p.Amount,
		Fees:            p.Fees,
		Description:     desc,
		ExpenseDatetime: p.Date,
		ReceiptType:     rt,
	}
	if err := s.repo.Create(e, p.WalletID, p.WalletBankDebitCardID); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *ExpenseService) FindAllWithCategory(filters domain.ExpenseFilters) ([]*domain.ExpenseWithCategory, error) {
	return s.repo.FindAllWithCategory(s.userID, filters)
}

func (s *ExpenseService) FindByID(id string) (*domain.Expense, error) {
	return s.repo.FindByID(s.userID, id)
}

func (s *ExpenseService) Update(e *domain.Expense) error {
	return s.repo.Update(e)
}

func (s *ExpenseService) Delete(id string) error {
	return s.repo.SoftDelete(s.userID, id)
}

func (s *ExpenseService) MonthlySummary(year, month int) (*domain.MonthlySummary, error) {
	return s.repo.MonthlySummary(s.userID, year, month)
}

func (s *ExpenseService) MonthlyTotal(year, month int) (float64, error) {
	return s.repo.MonthlyTotal(s.userID, year, month)
}
