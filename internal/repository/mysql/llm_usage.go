package mysql

import (
	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/jmoiron/sqlx"
)

type LLMUsageRepository struct {
	db     *sqlx.DB
	userID string
}

func NewLLMUsageRepository(db *sqlx.DB) *LLMUsageRepository {
	return &LLMUsageRepository{db: db, userID: domain.DefaultUserID}
}

func (r *LLMUsageRepository) ForUser(userID string) *LLMUsageRepository {
	copy := *r
	copy.userID = userID
	return &copy
}

// Log records a single LLM call. receiptType is optional ("image" | "pdf" | "").
func (r *LLMUsageRepository) Log(provider, model string, promptTokens, outputTokens int, purpose string, receiptType ...string) error {
	rt := ""
	if len(receiptType) > 0 {
		rt = receiptType[0]
	}
	var rtPtr *string
	if rt == "image" || rt == "pdf" {
		rtPtr = &rt
	}

	costUSD := llm.CalculateCost(model, promptTokens, outputTokens)

	_, err := r.db.Exec(`
		INSERT INTO llm_usage (user_id, provider, model, prompt_tokens, output_tokens, purpose, receipt_type, cost_usd)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, r.userID, provider, model, promptTokens, outputTokens, purpose, rtPtr, costUSD)
	return err
}
