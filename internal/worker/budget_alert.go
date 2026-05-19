package worker

import (
	"fmt"
	"html"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

// BudgetAlerter checks budget thresholds and sends Telegram alerts with inline buttons.
type BudgetAlerter struct {
	bot        *tgbotapi.BotAPI
	chatID     int64
	budgetSvc  *service.BudgetService
	settingSvc *service.SettingService
	// sent deduplicates alerts per category/month/threshold.
	// Key: "categoryID:year:month:threshold"
	sent map[string]bool
}

func NewBudgetAlerter(bot *tgbotapi.BotAPI, chatID int64, budgetSvc *service.BudgetService, settingSvc *service.SettingService) *BudgetAlerter {
	return &BudgetAlerter{
		bot:        bot,
		chatID:     chatID,
		budgetSvc:  budgetSvc,
		settingSvc: settingSvc,
		sent:       map[string]bool{},
	}
}

// Check is called after every expense save and by the periodic scheduler.
func (a *BudgetAlerter) Check() {
	now := time.Now()
	statuses, err := a.budgetSvc.StatusForMonth(now.Year(), int(now.Month()))
	if err != nil {
		log.Error().Err(err).Msg("budget-alert: failed to fetch statuses")
		return
	}

	for _, s := range statuses {
		a.maybeAlert(s)
	}
}

func (a *BudgetAlerter) maybeAlert(s *domain.BudgetStatus) {
	key := catKey(s)
	baseKey := fmt.Sprintf("%s:%d:%d", key, s.Year, s.Month)

	catLabel := "Overall"
	if s.CategoryName != "" {
		catLabel = s.CategoryIcon + " " + s.CategoryName
	}

	// Read threshold and enabled flag from settings (defaults: 80%, enabled).
	threshold := 80.0
	enabled := true
	if a.settingSvc != nil {
		threshold = a.settingSvc.GetFloat("budget_alert_threshold", 80)
		enabled = a.settingSvc.GetBool("budget_alert_enabled", true)
	}
	if !enabled {
		return
	}
	thresholdKey := fmt.Sprintf(":%d", int(threshold))

	switch {
	case s.Pct >= 100 && !a.sent[baseKey+":100"]:
		a.sent[baseKey+":100"] = true
		a.sendAlert(catLabel, s, 100)
	case s.Pct >= threshold && s.Pct < 100 && !a.sent[baseKey+thresholdKey]:
		a.sent[baseKey+thresholdKey] = true
		a.sendAlert(catLabel, s, 80)
	}
}

func (a *BudgetAlerter) sendAlert(catLabel string, s *domain.BudgetStatus, threshold int) {
	var emoji, verb string
	if threshold >= 100 {
		emoji = "🚨"
		verb = "exceeded"
	} else {
		emoji = "⚠️"
		verb = fmt.Sprintf("reached %d%% of", threshold)
	}

	currency := "USD"
	if a.settingSvc != nil {
		currency = a.settingSvc.GetString("currency", "USD")
	}

	text := fmt.Sprintf(
		"%s <b>Budget Alert — %s</b>\n\nYou've %s your <b>%s</b> budget.\n\n💰 Spent: <b>%.2f %s</b>\n📊 Budget: <b>%.2f %s</b>\n📈 Usage: <b>%.1f%%</b>",
		emoji, html.EscapeString(catLabel), verb, html.EscapeString(catLabel),
		s.Spent, currency, s.Effective, currency, s.Pct,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 View Summary", "summary"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Update Budget", "budget_set:"+catKey(s)),
		),
	)

	msg := tgbotapi.NewMessage(a.chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	if _, err := a.bot.Send(msg); err != nil {
		log.Error().Err(err).Str("category", catLabel).Int("threshold", threshold).Msg("budget-alert: send failed")
	}
}

func catKey(s *domain.BudgetStatus) string {
	if s.CategoryID != nil {
		return *s.CategoryID
	}
	return "overall"
}
