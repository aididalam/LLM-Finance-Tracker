package domain

import "time"

// Setting is one settings row per user.
type Setting struct {
	UserID               string    `db:"user_id"                json:"user_id"`
	Currency             string    `db:"currency"               json:"currency"`
	BudgetAlertThreshold int       `db:"budget_alert_threshold" json:"budget_alert_threshold"`
	BudgetAlertEnabled   bool      `db:"budget_alert_enabled"   json:"budget_alert_enabled"`
	Timezone             string    `db:"timezone"               json:"timezone"`
	Locale               string    `db:"locale"                 json:"locale"`
	TelegramChatID       *string   `db:"telegram_chat_id"       json:"telegram_chat_id,omitempty"`
	Metadata             *string   `db:"metadata"               json:"metadata,omitempty"`
	UpdatedAt            time.Time `db:"updated_at"             json:"updated_at"`
}

// SettingRepository defines the persistence contract for settings.
type SettingRepository interface {
	GetAll(userID string) (map[string]string, error)
	Set(userID, key, value string) error
}
