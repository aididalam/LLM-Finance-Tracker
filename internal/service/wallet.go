package service

import (
	"fmt"
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/google/uuid"
)

// WalletService manages wallets, sub-accounts, and internal transactions.
type WalletService struct {
	userID string
	repo   domain.WalletRepository
}

func NewWalletService(repo domain.WalletRepository) *WalletService {
	return &WalletService{repo: repo, userID: domain.DefaultUserID}
}

func (s *WalletService) ForUser(userID string) *WalletService {
	scoped := *s
	scoped.userID = userID
	return &scoped
}

// ── Wallet CRUD ──────────────────────────────────────────────────────────────

func (s *WalletService) ListWithBalances() ([]*domain.WalletWithBalance, error) {
	return s.repo.ListWithBalances(s.userID)
}

func (s *WalletService) Get(id string) (*domain.Wallet, error) {
	return s.repo.FindByID(s.userID, id)
}

func (s *WalletService) CreateWallet(name, accountType string) (*domain.Wallet, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !validWalletAccountType(accountType) {
		return nil, fmt.Errorf("invalid account_type: must be cash, bank, or mfs")
	}
	w := &domain.Wallet{
		ID:          uuid.New().String(),
		UserID:      s.userID,
		Name:        strings.TrimSpace(name),
		AccountType: accountType,
		IsActive:    true,
	}
	if err := s.repo.Create(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WalletService) UpdateWallet(id, name string) error {
	w, err := s.repo.FindByID(s.userID, id)
	if err != nil {
		return err
	}
	if w == nil {
		return fmt.Errorf("wallet not found")
	}
	w.Name = strings.TrimSpace(name)
	return s.repo.Update(w)
}

func (s *WalletService) DeactivateWallet(id string) error {
	return s.repo.Deactivate(s.userID, id)
}

func (s *WalletService) GetBalance(walletID string) (float64, error) {
	return s.repo.GetBalance(s.userID, walletID)
}

// ── Debit cards ──────────────────────────────────────────────────────────────

// GetDebitCards returns all debit cards for a bank wallet.
func (s *WalletService) GetDebitCards(walletID string) ([]*domain.WalletDebitCard, error) {
	return s.repo.GetDebitCards(walletID)
}

// AddDebitCard upserts the bank details then creates a new debit card entry.
func (s *WalletService) AddDebitCard(walletID, bankName, accountNumber, last4Digit string, branch *string) (*domain.WalletDebitCard, error) {
	w, err := s.repo.FindByID(s.userID, walletID)
	if err != nil || w == nil {
		return nil, fmt.Errorf("wallet not found")
	}
	if w.AccountType != "bank" {
		return nil, fmt.Errorf("debit cards can only be added to bank accounts")
	}

	// Ensure a wallet_bank record exists (upsert with provided details).
	existing, err := s.repo.GetBankDetails(walletID)
	if err != nil {
		return nil, err
	}
	var bankID string
	if existing != nil {
		bankID = existing.ID
	} else {
		bankID = uuid.New().String()
		if err := s.repo.UpsertBankDetails(&domain.WalletBank{
			ID:            bankID,
			WalletID:      walletID,
			IsActive:      true,
			BankName:      bankName,
			Branch:        branch,
			AccountNumber: accountNumber,
		}); err != nil {
			return nil, err
		}
	}

	card := &domain.WalletDebitCard{
		ID:           uuid.New().String(),
		WalletBankID: bankID,
		WalletID:     walletID,
		IsActive:     true,
		Last4Digit:   last4Digit,
	}
	if err := s.repo.AddDebitCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

// ── Internal transactions ─────────────────────────────────────────────────────

// AddInitialBalance records the starting balance for a wallet.
func (s *WalletService) AddInitialBalance(walletID string, amount, fees float64) error {
	return s.createInternalTx(walletID, "initial", nil, amount, fees)
}

// AddIncome records income into a wallet.
func (s *WalletService) AddIncome(walletID string, amount, fees float64) error {
	return s.createInternalTx(walletID, "income", nil, amount, fees)
}

// Transfer moves funds between wallets.
func (s *WalletService) Transfer(fromWalletID, toWalletID string, amount, fees float64) error {
	if fromWalletID == toWalletID {
		return fmt.Errorf("source and destination wallets must be different")
	}
	from, err := s.repo.FindByID(s.userID, fromWalletID)
	if err != nil || from == nil {
		return fmt.Errorf("source wallet not found")
	}
	to, err := s.repo.FindByID(s.userID, toWalletID)
	if err != nil || to == nil {
		return fmt.Errorf("destination wallet not found")
	}
	return s.createInternalTx(fromWalletID, "internal", &toWalletID, amount, fees)
}

func (s *WalletService) createInternalTx(walletID, sourceType string, toWallet *string, amount, fees float64) error {
	return s.repo.CreateInternalTransaction(&domain.WalletInternalTransaction{
		ID:         uuid.New().String(),
		WalletID:   walletID,
		SourceType: sourceType,
		ToWalletID: toWallet,
		Amount:     amount,
		Fees:       fees,
	})
}

// MonthlyIncomeTotal returns the sum of income transactions for a given month.
func (s *WalletService) MonthlyIncomeTotal(year, month int) (float64, error) {
	return s.repo.MonthlyIncomeTotal(s.userID, year, month)
}

// DefaultWalletID returns the ID of the first active wallet for the user (cash first),
// or "" if none exist.
func (s *WalletService) DefaultWalletID() (string, error) {
	wallets, err := s.repo.List(s.userID)
	if err != nil {
		return "", err
	}
	if len(wallets) == 0 {
		return "", nil
	}
	return wallets[0].ID, nil
}

func validWalletAccountType(t string) bool {
	switch t {
	case "cash", "bank", "mfs":
		return true
	default:
		return false
	}
}
