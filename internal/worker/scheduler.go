package worker

import (
	"fmt"
	"html"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

func StartScheduler(bot *tgbotapi.BotAPI, chatID int64, expenseSvc *service.ExpenseService, alerter *BudgetAlerter) (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	// Weekly digest — every Monday at 09:00
	_, err = s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday), gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() { sendWeeklyDigest(bot, chatID, expenseSvc) }),
	)
	if err != nil {
		return nil, fmt.Errorf("register weekly digest: %w", err)
	}

	// Monthly report — 1st of every month at 09:00
	_, err = s.NewJob(
		gocron.MonthlyJob(1, gocron.NewDaysOfTheMonth(1), gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() { sendMonthlyReport(bot, chatID, expenseSvc) }),
	)
	if err != nil {
		return nil, fmt.Errorf("register monthly report: %w", err)
	}

	// Budget alert check — every hour
	_, err = s.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(alerter.Check),
	)
	if err != nil {
		return nil, fmt.Errorf("register budget alert job: %w", err)
	}

	s.Start()
	log.Info().Msg("scheduler: started (weekly digest, monthly report, budget alerts)")
	return s, nil
}

func domainFilters(from, to time.Time) domain.ExpenseFilters {
	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")
	return domain.ExpenseFilters{From: &fromStr, To: &toStr, Limit: 1000}
}

func sendWeeklyDigest(bot *tgbotapi.BotAPI, chatID int64, expenseSvc *service.ExpenseService) {
	now := time.Now()
	from := now.AddDate(0, 0, -7)
	filters := domainFilters(from, now)
	expenses, err := expenseSvc.FindAllWithCategory(filters)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: weekly digest query failed")
		return
	}

	var total float64
	for _, e := range expenses {
		total += e.Amount
	}

	text := fmt.Sprintf("📅 <b>Weekly Digest</b>\n\n<i>%s — %s</i>\n\n🔴 Expenses: <b>%.2f BDT</b>\n📦 Transactions: <b>%d</b>\n\nUse /summary for the full monthly breakdown.",
		html.EscapeString(from.Format("02 Jan")), html.EscapeString(now.Format("02 Jan 2006")), total, len(expenses))

	send(bot, chatID, text)
}

func sendMonthlyReport(bot *tgbotapi.BotAPI, chatID int64, expenseSvc *service.ExpenseService) {
	prev := time.Now().AddDate(0, -1, 0)
	expenses, err := expenseSvc.MonthlySummary(prev.Year(), int(prev.Month()))
	if err != nil {
		log.Error().Err(err).Msg("scheduler: monthly report failed")
		return
	}

	text := fmt.Sprintf("📊 <b>Monthly Report — %s</b>\n\n🔴 Expenses: <b>%.2f BDT</b>\n",
		html.EscapeString(prev.Format("January 2006")), expenses.Total)

	if len(expenses.Categories) > 0 {
		text += "\nExpense categories:\n"
		for _, c := range expenses.Categories {
			text += fmt.Sprintf("%s <b>%s</b>: %.2f BDT (%d)\n", c.Icon, html.EscapeString(c.CategoryName), c.Total, c.Count)
		}
	} else {
		text += "\nNo expenses recorded.\n"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Export CSV", "export_csv"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("scheduler: monthly report send failed")
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := bot.Send(msg); err != nil {
		log.Error().Err(err).Msg("scheduler: send failed")
	}
}
