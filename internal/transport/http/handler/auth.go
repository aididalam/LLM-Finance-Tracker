package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aididalam/llmexpensetracker/internal/auth"
	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/jmoiron/sqlx"
)

type AuthHandler struct {
	cfg *config.Config
	db  *sqlx.DB
}

func NewAuthHandler(cfg *config.Config, db *sqlx.DB) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: db}
}

func (h *AuthHandler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	var data auth.TelegramAuthData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.ResError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := data.Verify(h.cfg.TelegramBotToken); err != nil {
		response.ResError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	allowedID, _ := strconv.ParseInt(h.cfg.TelegramChatID, 10, 64)
	if data.ID != allowedID {
		response.ResError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Look up the user_id associated with this telegram chat id
	userID := h.lookupUserByTelegramChatID(strconv.FormatInt(data.ID, 10))

	token, err := auth.IssueToken(data.ID, userID, h.cfg.JWTSecret)
	if err != nil {
		response.ResError(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	response.ResSuccess(w, map[string]string{"token": token})
}

// lookupUserByTelegramChatID returns the user_id for a telegram chat ID,
// falling back to DefaultUserID if no mapping exists.
func (h *AuthHandler) lookupUserByTelegramChatID(chatID string) string {
	if h.db == nil {
		return domain.DefaultUserID
	}
	var userID string
	err := h.db.QueryRow(
		"SELECT user_id FROM users_settings WHERE telegram_chat_id = ? LIMIT 1",
		chatID,
	).Scan(&userID)
	if err == sql.ErrNoRows || userID == "" {
		return domain.DefaultUserID
	}
	if err != nil {
		return domain.DefaultUserID
	}
	return userID
}

// BotName fetches the bot's username from Telegram so the login widget can load it dynamically.
func (h *AuthHandler) BotName(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getMe", h.cfg.TelegramBotToken))
	if err != nil || resp.StatusCode != 200 {
		response.ResError(w, "failed to fetch bot info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tgResp struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil || !tgResp.OK {
		response.ResError(w, "failed to parse bot info", http.StatusInternalServerError)
		return
	}

	response.ResSuccess(w, map[string]string{"bot_name": tgResp.Result.Username})
}
