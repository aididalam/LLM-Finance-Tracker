package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type settingRepository struct {
	db *sqlx.DB
}

// NewSettingRepository creates a new MySQL-backed setting repository.
func NewSettingRepository(db *sqlx.DB) *settingRepository {
	return &settingRepository{db: db}
}

// GetAll returns all settings as a map[key]value.
func (r *settingRepository) GetAll(userID string) (map[string]string, error) {
	if err := r.ensureRow(userID); err != nil {
		return nil, err
	}

	var s struct {
		UserID               string         `db:"user_id"`
		Currency             string         `db:"currency"`
		BudgetAlertThreshold int            `db:"budget_alert_threshold"`
		BudgetAlertEnabled   bool           `db:"budget_alert_enabled"`
		Timezone             string         `db:"timezone"`
		Locale               string         `db:"locale"`
		TelegramChatID       sql.NullString `db:"telegram_chat_id"`
		Metadata             sql.NullString `db:"metadata"`
	}
	err := r.db.Get(&s, `
		SELECT user_id, currency, budget_alert_threshold, budget_alert_enabled,
		       timezone, locale, telegram_chat_id, metadata
		FROM users_settings
		WHERE user_id = ?
		LIMIT 1
	`, userID)
	if err != nil {
		return nil, err
	}

	result := map[string]string{
		"currency":               s.Currency,
		"budget_alert_threshold": strconv.Itoa(s.BudgetAlertThreshold),
		"budget_alert_enabled":   strconv.FormatBool(s.BudgetAlertEnabled),
		"timezone":               s.Timezone,
		"locale":                 s.Locale,
	}
	if s.TelegramChatID.Valid {
		result["telegram_chat_id"] = s.TelegramChatID.String
	}
	if s.Metadata.Valid {
		result["metadata"] = s.Metadata.String
	}
	return result, nil
}

// Set inserts or updates a setting value.
func (r *settingRepository) Set(userID, key, value string) error {
	column, ok := settingColumn(key)
	if !ok {
		return nil
	}
	normalized, err := normalizeSettingValue(key, value)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(
		"INSERT INTO users_settings (user_id, %s) VALUES (?, ?) ON DUPLICATE KEY UPDATE %s = VALUES(%s)",
		column, column, column,
	)
	_, err = r.db.Exec(
		query,
		userID, normalized,
	)
	return err
}

func (r *settingRepository) ensureRow(userID string) error {
	_, err := r.db.Exec("INSERT IGNORE INTO users_settings (user_id) VALUES (?)", userID)
	return err
}

func normalizeSettingValue(key, value string) (any, error) {
	v := strings.TrimSpace(value)
	switch key {
	case "currency":
		if v == "" {
			v = "USD"
		}
		return strings.ToUpper(v), nil
	case "budget_alert_threshold":
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			n = 0
		}
		if n > 100 {
			n = 100
		}
		return n, nil
	case "budget_alert_enabled":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		return b, nil
	case "timezone":
		if v == "" {
			v = "Asia/Dhaka"
		}
		return v, nil
	case "locale":
		if v == "" {
			v = "en-BD"
		}
		return v, nil
	case "telegram_chat_id":
		if v == "" {
			return nil, nil
		}
		return v, nil
	case "metadata":
		if v == "" {
			return nil, nil
		}
		if !json.Valid([]byte(v)) {
			return nil, fmt.Errorf("metadata must be valid JSON")
		}
		return v, nil
	default:
		return v, nil
	}
}

func settingColumn(key string) (string, bool) {
	switch key {
	case "currency":
		return "currency", true
	case "budget_alert_threshold":
		return "budget_alert_threshold", true
	case "budget_alert_enabled":
		return "budget_alert_enabled", true
	case "timezone":
		return "timezone", true
	case "locale":
		return "locale", true
	case "telegram_chat_id":
		return "telegram_chat_id", true
	case "metadata":
		return "metadata", true
	default:
		return "", false
	}
}

// LookupUserByTelegramChatID finds the user_id associated with a telegram chat ID.
func (r *settingRepository) LookupUserByTelegramChatID(chatID string) (string, error) {
	var userID string
	err := r.db.QueryRow(
		"SELECT user_id FROM users_settings WHERE telegram_chat_id = ? LIMIT 1",
		chatID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return userID, err
}
