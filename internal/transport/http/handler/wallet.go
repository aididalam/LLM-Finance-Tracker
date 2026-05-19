package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
)

type AccountHandler struct {
	svc *service.WalletService
}

func NewAccountHandler(svc *service.WalletService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	wallets, err := svc.ListWithBalances()
	if err != nil {
		response.ResError(w, "failed to fetch wallets", http.StatusInternalServerError)
		return
	}
	if wallets == nil {
		wallets = []*domain.WalletWithBalance{}
	}
	response.ResSuccess(w, wallets)
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		AccountType string `json:"account_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		response.ResError(w, "name is required", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	wallet, err := svc.CreateWallet(body.Name, body.AccountType)
	if err != nil {
		response.ResError(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.ResSuccess(w, wallet, http.StatusCreated)
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		response.ResError(w, "name is required", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	if err := svc.UpdateWallet(id, body.Name); err != nil {
		response.ResError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, nil)
}

func (h *AccountHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	if err := svc.DeactivateWallet(id); err != nil {
		response.ResError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, nil)
}

// Transfer records an internal funds transfer between two wallets.
func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromWalletID string  `json:"from_wallet_id"`
		ToWalletID   string  `json:"to_wallet_id"`
		Amount       float64 `json:"amount"`
		Fees         float64 `json:"fees"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.FromWalletID == "" || body.ToWalletID == "" {
		response.ResError(w, "from_wallet_id and to_wallet_id are required", http.StatusBadRequest)
		return
	}
	if body.Amount <= 0 {
		response.ResError(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	if err := svc.Transfer(body.FromWalletID, body.ToWalletID, body.Amount, body.Fees); err != nil {
		response.ResError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, nil)
}

// AddIncome records an income transaction for a wallet.
func (h *AccountHandler) AddIncome(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Amount float64 `json:"amount"`
		Fees   float64 `json:"fees"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Amount <= 0 {
		response.ResError(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	if err := svc.AddIncome(id, body.Amount, body.Fees); err != nil {
		response.ResError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, nil, http.StatusCreated)
}

// AddInitialBalance seeds the starting balance for a wallet.
func (h *AccountHandler) AddInitialBalance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Amount float64 `json:"amount"`
		Fees   float64 `json:"fees"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	if err := svc.AddInitialBalance(id, body.Amount, body.Fees); err != nil {
		response.ResError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, nil, http.StatusCreated)
}

// ListDebitCards returns all debit cards for a bank account.
//
//	GET /accounts/{id}/debit-cards
func (h *AccountHandler) ListDebitCards(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	cards, err := svc.GetDebitCards(id)
	if err != nil {
		response.ResError(w, "failed to fetch cards", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, cards)
}

// AddDebitCard adds a debit card to a bank account, creating the bank record if needed.
//
//	POST /accounts/{id}/debit-cards
func (h *AccountHandler) AddDebitCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Last4Digit    string  `json:"last_4_digit"`
		BankName      string  `json:"bank_name"`
		AccountNumber string  `json:"account_number"`
		Branch        *string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.Last4Digit) != 4 {
		response.ResError(w, "last_4_digit must be exactly 4 digits", http.StatusBadRequest)
		return
	}
	svc := h.svc.ForUser(middleware.CurrentUserID(r.Context()))
	card, err := svc.AddDebitCard(id, body.BankName, body.AccountNumber, body.Last4Digit, body.Branch)
	if err != nil {
		response.ResError(w, err.Error(), http.StatusBadRequest)
		return
	}
	response.ResSuccess(w, card, http.StatusCreated)
}
