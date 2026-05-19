package telegram

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

// ── Message routing ───────────────────────────────────────────────────────────

func (h *Handler) send(chatID int64, text string) {
	out := tgbotapi.NewMessage(chatID, text)
	out.ParseMode = "HTML"
	if _, err := h.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("bot: failed to send message")
	}
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	if h.budgetPrompt != nil && !msg.IsCommand() && msg.Text != "" {
		h.handleBudgetAmountReply(msg)
		return
	}
	if msg.IsCommand() {
		h.handleCommand(msg)
		return
	}
	if len(msg.Photo) > 0 {
		best := msg.Photo[len(msg.Photo)-1]
		if msg.MediaGroupID != "" {
			if h.mediaGroupBuf == nil {
				h.mediaGroupBuf = make(map[string][]string)
			}
			h.mediaGroupBuf[msg.MediaGroupID] = append(h.mediaGroupBuf[msg.MediaGroupID], best.FileID)
			return
		}
		h.handleReceipt(msg.Chat.ID, best.FileID, "image/jpeg")
		return
	}
	if msg.Document != nil {
		if receiptMIME(msg.Document.MimeType) {
			h.handleReceipt(msg.Chat.ID, msg.Document.FileID, msg.Document.MimeType)
			return
		}
		h.send(msg.Chat.ID, "Please send a receipt as a photo or PDF document.")
		return
	}
	if msg.Text == "" {
		h.send(msg.Chat.ID, "Send me a photo or PDF of a receipt, or describe a transaction in text.")
		return
	}
	h.parseExpense(msg)
}

// ── Decision routing ──────────────────────────────────────────────────────────
// handleDecision maps a ChatDecision from ChatProcessor to Telegram UI.
// Mirrors how the HTTP handler maps ChatDecision to a JSON response.

func (h *Handler) handleDecision(chatID int64, decision service.ChatDecision, receiptType ...string) {
	rt := ""
	if len(receiptType) > 0 {
		rt = receiptType[0]
	}
	switch decision.Action {
	case service.ChatActionConfirm:
		h.clearHistory()
		h.presentParsedConfirmation(chatID, decision.Expense, rt, "")
	case service.ChatActionClarify:
		h.addToHistory("assistant", decision.Reply)
		h.send(chatID, html.EscapeString(decision.Reply))
	case service.ChatActionDeleteSelect:
		h.clearHistory()
		h.showDeleteMatches(chatID, decision.Expenses, decision.Reply)
	default: // ChatActionChat, ChatActionAnswer
		h.clearHistory()
		h.send(chatID, html.EscapeString(decision.Reply))
	}
}

// ── Receipt ───────────────────────────────────────────────────────────────────

func receiptMIME(mime string) bool {
	switch mime {
	case "application/pdf", "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

func (h *Handler) downloadFile(fileID string) ([]byte, error) {
	tgFile, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(tgFile.Link(h.bot.Token)) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 20<<20))
}

func (h *Handler) flushMediaGroups() {
	if len(h.mediaGroupBuf) == 0 {
		return
	}
	for groupID, fileIDs := range h.mediaGroupBuf {
		delete(h.mediaGroupBuf, groupID)
		if len(fileIDs) == 1 {
			h.handleReceipt(h.chatID, fileIDs[0], "image/jpeg")
		} else {
			h.handleMultiPhotoAlbum(h.chatID, fileIDs)
		}
	}
}

func (h *Handler) handleMultiPhotoAlbum(chatID int64, fileIDs []string) {
	h.send(chatID, fmt.Sprintf("📷 Processing %d photos…", len(fileIDs)))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	categories, err := h.categorySvc.Names()
	if err != nil {
		h.send(chatID, "Something went wrong. Please try again.")
		return
	}

	var allItems []llm.ReceiptItem
	for _, fileID := range fileIDs {
		data, err := h.downloadFile(fileID)
		if err != nil {
			log.Error().Err(err).Str("file_id", fileID).Msg("bot: album photo download failed")
			continue
		}
		parsed, usage, err := h.llmProvider.ParseReceipt(ctx, data, "image/jpeg", categories, h.userCurrency())
		if usage != nil {
			_ = h.llmUsage.Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "receipt", "image")
		}
		if err != nil || parsed == nil || !parsed.IsExpense {
			continue
		}
		parsed.Normalize()
		if len(parsed.Items) > 0 {
			allItems = append(allItems, parsed.Items...)
		} else {
			allItems = append(allItems, llm.ReceiptItem{
				Description:   parsed.Description,
				Amount:        parsed.Amount,
				Currency:      parsed.Currency,
				Category:      parsed.Category,
				Subcategory:   parsed.Subcategory,
				PaymentMethod: parsed.PaymentMethod,
				ExpenseDate:   parsed.ExpenseDate,
			})
		}
	}

	if len(allItems) == 0 {
		h.send(chatID, "Could not parse any receipts from the photos.")
		return
	}
	h.handleMultiItemReceipt(chatID, &llm.ParsedExpense{IsExpense: true, Items: allItems}, "image")
}

func (h *Handler) handleReceipt(chatID int64, fileID, mediaType string) {
	h.send(chatID, "📄 Reading your receipt…")

	data, err := h.downloadFile(fileID)
	if err != nil {
		log.Error().Err(err).Msg("bot: receipt download failed")
		h.send(chatID, "Could not download the file. Please try again.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	categories, err := h.categorySvc.Names()
	if err != nil {
		h.send(chatID, "Something went wrong. Please try again.")
		return
	}

	parsed, usage, err := h.llmProvider.ParseReceipt(ctx, data, mediaType, categories, h.userCurrency())
	rt := "image"
	if mediaType == "application/pdf" {
		rt = "pdf"
	}
	if err != nil {
		log.Error().Err(err).Msg("bot: ParseReceipt failed")
		msg := err.Error()
		if len(msg) > 10 && len(msg) < 300 {
			h.send(chatID, "❌ "+html.EscapeString(msg))
		} else {
			h.send(chatID, "Could not parse the receipt. Try sending a clearer image.")
		}
		return
	}
	if usage != nil {
		_ = h.llmUsage.Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "receipt", rt)
	}

	parsed.Normalize()

	if parsed.MovementType != "" && parsed.MovementType != "transaction" {
		h.handleDecision(chatID, h.newChatProcessor().ProcessParsed(parsed), rt)
		return
	}

	if !parsed.IsExpense {
		reply := parsed.NotExpenseReply
		if reply == "" {
			reply = "Could not read this receipt clearly. Please try a clearer photo."
		}
		h.send(chatID, html.EscapeString(reply))
		return
	}

	if len(parsed.Items) > 1 {
		h.handleMultiItemReceipt(chatID, parsed, rt)
		return
	}

	h.autoSavePending(chatID)
	cat, _ := h.categorySvc.FindOrCreate(parsed.Category)
	var catID *string
	if cat != nil {
		id := cat.ID
		catID = &id
	}
	h.pending = &pending{Parsed: parsed, CategoryID: catID, ReceiptType: rt}
	h.presentParsedConfirmation(chatID, parsed, rt, "Receipt Parsed")
}

func (h *Handler) handleMultiItemReceipt(chatID int64, parsed *llm.ParsedExpense, receiptType string) {
	h.autoSavePending(chatID)

	for _, item := range parsed.Items {
		if item.ToParsed(h.userCurrency()).MovementType == "ambiguous_card" {
			// ambiguous card prompts no longer supported — just proceed with save
			break
		}
	}

	var sb strings.Builder
	sb.WriteString("🧾 <b>Receipt — Multiple Items</b>\n\n")
	var total float64
	for i, item := range parsed.Items {
		fmt.Fprintf(&sb, "%d. %s — <b>%.2f %s</b>\n",
			i+1, html.EscapeString(item.Description), item.Amount, item.Currency)
		total += item.Amount
	}
	cur := h.userCurrency()
	if len(parsed.Items) > 0 && parsed.Items[0].Currency != "" {
		cur = parsed.Items[0].Currency
	}
	fmt.Fprintf(&sb, "\n<b>Total: %.2f %s</b>\n\nSave all together or confirm one by one?", total, cur)

	h.pending = &pending{Items: parsed.Items, ReceiptType: receiptType}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Save All", "receipt_save_all"),
			tgbotapi.NewInlineKeyboardButtonData("📋 One by One", "receipt_one_by_one"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel"),
		),
	)
	out := tgbotapi.NewMessage(chatID, sb.String())
	out.ParseMode = "HTML"
	out.ReplyMarkup = keyboard
	h.bot.Send(out) //nolint:errcheck
}

func (h *Handler) showNextItem(chatID int64) {
	if h.pending == nil || h.pending.ItemIndex >= len(h.pending.Items) {
		h.pending = nil
		h.send(chatID, "✅ All receipt items processed!")
		return
	}
	item := h.pending.Items[h.pending.ItemIndex]
	p := item.ToParsed(h.userCurrency())

	date, _ := time.Parse("2006-01-02", p.ExpenseDate)
	dateStr := date.Format("02 Jan 2006")
	if dateStr == "01 Jan 0001" {
		dateStr = time.Now().Format("02 Jan 2006")
	}
	cat, _ := h.categorySvc.FindOrCreate(p.Category)
	catName := p.Category
	if cat != nil {
		catName = fmt.Sprintf("%s %s", cat.Icon, cat.Name)
	}
	if p.Subcategory != "" {
		catName = fmt.Sprintf("%s › %s", catName, p.Subcategory)
	}

	text := fmt.Sprintf(
		"📋 Item %d/%d\n\n💰 Amount: <b>%.2f %s</b>\n📂 Category: %s\n📝 Description: %s\n📅 Date: %s\n\nConfirm to save?",
		h.pending.ItemIndex+1, len(h.pending.Items),
		p.Amount, p.Currency,
		html.EscapeString(catName),
		html.EscapeString(p.Description),
		dateStr,
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "confirm"),
			tgbotapi.NewInlineKeyboardButtonData("⏭ Skip", "receipt_skip"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel All", "cancel"),
		),
	)
	out := tgbotapi.NewMessage(chatID, text)
	out.ParseMode = "HTML"
	out.ReplyMarkup = keyboard
	h.bot.Send(out) //nolint:errcheck
}

func (h *Handler) saveCurrentReceiptItem(chatID int64) {
	item := h.pending.Items[h.pending.ItemIndex]
	p := item.ToParsed(h.userCurrency())
	cat, _ := h.categorySvc.FindOrCreate(p.Category)
	var catID *string
	if cat != nil {
		id := cat.ID
		catID = &id
	}
	walletID := h.pending.WalletID
	if walletID == "" {
		walletID = h.defaultWalletID()
	}
	if _, _, err := h.expenseSvc.CreateFromParsed(p, catID, walletID, h.pending.ReceiptType); err != nil {
		log.Error().Err(err).Msg("bot: failed to save receipt item")
		h.send(chatID, "❌ Failed to save item. Skipping.")
	}
	h.pending.ItemIndex++
	if h.pending.ItemIndex >= len(h.pending.Items) {
		total := len(h.pending.Items)
		h.pending = nil
		h.send(chatID, fmt.Sprintf("✅ All %d items saved!", total))
	} else {
		h.showNextItem(chatID)
	}
}

// ── Conversation flow ─────────────────────────────────────────────────────────

func (h *Handler) parseExpense(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	categories, err := h.categorySvc.Names()
	if err != nil {
		log.Error().Err(err).Msg("bot: failed to fetch categories")
		h.send(msg.Chat.ID, "Something went wrong. Please try again.")
		return
	}

	h.addToHistory("user", msg.Text)

	decision, usage, err := h.newChatProcessor().ProcessMessages(ctx, h.conversationHistory, categories)
	if err != nil {
		log.Error().Err(err).Msg("bot: LLM chat failed")
		h.send(msg.Chat.ID, "Couldn't parse that. Please try again.")
		return
	}
	if usage != nil {
		_ = h.llmUsage.Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "chat")
	}

	h.handleDecision(msg.Chat.ID, decision)
}

func (h *Handler) handleCardChoiceReply(msg *tgbotapi.Message) {
	// card prompts are no longer supported — just cancel
	h.send(msg.Chat.ID, "❌ Card payment method prompts are no longer supported.")
}

// handleLoanPromptReply — removed (loan flow no longer supported)

func (h *Handler) autoSavePending(chatID int64) {
	if h.pending == nil {
		return
	}
	if h.pending.Parsed == nil {
		h.pending = nil
		return
	}
	p := h.pending
	h.pending = nil
	walletID := p.WalletID
	if walletID == "" {
		walletID = h.defaultWalletID()
	}
	if _, _, err := h.expenseSvc.CreateFromParsed(p.Parsed, p.CategoryID, walletID, p.ReceiptType, p.WalletBankDebitCardID); err != nil {
		log.Error().Err(err).Msg("bot: auto-save pending failed")
		return
	}
	h.send(chatID, fmt.Sprintf("✅ Auto-saved expense: <b>%.2f %s</b> (%s)",
		p.Parsed.Amount, p.Parsed.Currency,
		html.EscapeString(p.Parsed.Description)))
}

func (h *Handler) presentParsedConfirmation(chatID int64, parsed *llm.ParsedExpense, receiptType, title string) {
	h.autoSavePending(chatID)

	cat, err := h.categorySvc.FindOrCreate(parsed.Category)
	if err != nil {
		h.send(chatID, "Something went wrong. Please try again.")
		return
	}
	var catID *string
	if cat != nil {
		id := cat.ID
		catID = &id
	}
	h.pending = &pending{Parsed: parsed, CategoryID: catID, ReceiptType: receiptType}

	date, _ := time.Parse("2006-01-02", parsed.ExpenseDate)
	dateStr := date.Format("02 Jan 2006")
	if dateStr == "01 Jan 0001" {
		dateStr = time.Now().Format("02 Jan 2006")
	}
	catName := parsed.Category
	if cat != nil {
		catName = fmt.Sprintf("%s %s", cat.Icon, cat.Name)
	}
	if parsed.Subcategory != "" {
		catName = fmt.Sprintf("%s › %s", catName, parsed.Subcategory)
	}
	payIcon := "💵"
	if parsed.PaymentMethod == "credit_card" || parsed.PaymentMethod == "debit_card" {
		payIcon = "💳"
	}
	txType := parsed.TransactionType
	if title == "" {
		title = transactionTitle(txType) + " Detected"
	}

	text := fmt.Sprintf(
		"%s <b>%s</b>\n\n💰 Amount: <b>%.2f %s</b>\n📂 Category: %s\n📝 Description: %s\n📅 Date: %s\n%s Payment: %s\n\nConfirm this %s?",
		transactionIcon(txType), html.EscapeString(title),
		parsed.Amount, parsed.Currency,
		html.EscapeString(catName),
		html.EscapeString(parsed.Description),
		dateStr, payIcon,
		html.EscapeString(strings.ReplaceAll(parsed.PaymentMethod, "_", " ")),
		transactionLabel(txType),
	)
	// Resolve wallet options based on payment method hint
	wallets, _ := h.walletSvc.ListWithBalances()
	matched := service.MatchWallets(parsed.PaymentMethod, wallets)

	if len(matched) == 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "confirm"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel"),
			),
		)
		out := tgbotapi.NewMessage(chatID, text)
		out.ParseMode = "HTML"
		out.ReplyMarkup = keyboard
		h.bot.Send(out) //nolint:errcheck
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, w := range matched {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				walletButtonLabel(w.Name, w.AccountType),
				fmt.Sprintf("wallet:%s:%s", w.ID, w.AccountType),
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel"),
	))
	out := tgbotapi.NewMessage(chatID, text+"\n\n💳 <b>Select wallet:</b>")
	out.ParseMode = "HTML"
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := h.bot.Send(out); err != nil {
		log.Error().Err(err).Msg("bot: failed to send confirmation")
	}
}

// confirmAndSave saves the pending expense using the selected wallet and debit card.
func (h *Handler) confirmAndSave(chatID int64) {
	p := h.pending
	h.pending = nil
	h.clearHistory()
	walletID := p.WalletID
	if walletID == "" {
		walletID = h.defaultWalletID()
	}
	if _, _, err := h.expenseSvc.CreateFromParsed(p.Parsed, p.CategoryID, walletID, p.ReceiptType, p.WalletBankDebitCardID); err != nil {
		log.Error().Err(err).Msg("bot: failed to save transaction")
		h.send(chatID, "Failed to save. Please try again.")
		return
	}
	h.send(chatID, "✅ Expense saved!")
}

// presentDebitCardKeyboard shows a debit card selection keyboard for a bank wallet.
func (h *Handler) presentDebitCardKeyboard(chatID int64) {
	if h.pending == nil {
		return
	}
	cards, err := h.walletSvc.GetDebitCards(h.pending.WalletID)
	if err != nil || len(cards) == 0 {
		h.confirmAndSave(chatID)
		return
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏦 No card", "card_none"),
	))
	for _, c := range cards {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("💳 **** %s", c.Last4Digit),
				"card:"+c.ID,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel"),
	))
	out := tgbotapi.NewMessage(chatID, "💳 <b>Select debit card:</b>")
	out.ParseMode = "HTML"
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(out) //nolint:errcheck
}

// walletButtonLabel returns a labelled button string for a wallet.
func walletButtonLabel(name, accountType string) string {
	switch accountType {
	case "cash":
		return "💵 " + name
	case "bank":
		return "🏦 " + name
	case "mfs":
		return "📱 " + name
	}
	return "💳 " + name
}

func (h *Handler) showDeleteMatches(chatID int64, expenses []*domain.ExpenseWithCategory, header string) {
	h.deleteMatches = expenses

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n\n", header)
	for i, e := range expenses {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		fmt.Fprintf(&sb, "%d. %s — <b>%.2f</b> (%s)\n",
			i+1, html.EscapeString(desc), e.Amount,
			e.ExpenseDatetime.Format("02 Jan"))
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, e := range expenses {
		desc := ""
		if e.Description != nil {
			desc = *e.Description
		}
		label := desc
		if len(label) > 28 {
			label = label[:25] + "…"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🗑 %s (%.0f)", label, e.Amount),
				"del:"+e.ID,
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "delete_done"),
	))

	out := tgbotapi.NewMessage(chatID, sb.String())
	out.ParseMode = "HTML"
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(out) //nolint:errcheck
}

// ── Commands & Callbacks ──────────────────────────────────────────────────────

func (h *Handler) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.send(msg.Chat.ID, "👋 Welcome! Send me your expenses or income and I'll log them.\n\nExamples:\n• <i>Spent 250 BDT on lunch</i>\n• <i>Got 50000 BDT salary</i>\n• <i>Coffee $4.50</i>\n\n/summary /budget /report /categories /help")
	case "help":
		h.send(msg.Chat.ID, "Just describe what you spent or received. I'll figure out the rest.\n\n/summary — this month's summary\n/budget — budget status\n/report — full monthly report\n/categories — list categories\n/help — this message")
	case "summary":
		h.sendSummary(msg.Chat.ID)
	case "budget":
		h.sendBudgetStatus(msg.Chat.ID)
	case "report":
		h.sendMonthlyReport(msg.Chat.ID)
	case "categories":
		h.sendCategories(msg.Chat.ID)
	}
}

func (h *Handler) handleCallback(cb *tgbotapi.CallbackQuery) {
	h.bot.Send(tgbotapi.NewCallback(cb.ID, "")) //nolint:errcheck

	data := cb.Data
	chatID := cb.Message.Chat.ID

	switch {
	case data == "confirm":
		if h.pending == nil {
			h.send(chatID, "No pending transaction.")
			return
		}
		if h.pending.OneByOne {
			h.saveCurrentReceiptItem(chatID)
			return
		}
		h.confirmAndSave(chatID)

	case strings.HasPrefix(data, "wallet:"):
		if h.pending == nil {
			return
		}
		parts := strings.SplitN(strings.TrimPrefix(data, "wallet:"), ":", 2)
		if len(parts) != 2 {
			return
		}
		h.pending.WalletID = parts[0]
		if parts[1] == "bank" {
			h.presentDebitCardKeyboard(chatID)
		} else {
			h.confirmAndSave(chatID)
		}

	case data == "card_none":
		if h.pending == nil {
			return
		}
		h.confirmAndSave(chatID)

	case strings.HasPrefix(data, "card:"):
		if h.pending == nil {
			return
		}
		cardID := strings.TrimPrefix(data, "card:")
		h.pending.WalletBankDebitCardID = &cardID
		h.confirmAndSave(chatID)

	case data == "cancel":
		h.pending = nil
		h.deleteMatches = nil
		h.clearHistory()
		h.send(chatID, "❌ Cancelled.")

	case data == "receipt_save_all":
		if h.pending == nil || len(h.pending.Items) == 0 {
			h.send(chatID, "No pending receipt items.")
			return
		}
		items := h.pending.Items
		rt := h.pending.ReceiptType
		h.pending = nil
		saved := 0
		for _, item := range items {
			p := item.ToParsed(h.userCurrency())
			cat, _ := h.categorySvc.FindOrCreate(p.Category)
			var catID *string
			if cat != nil {
				id := cat.ID
				catID = &id
			}
			if _, _, err := h.expenseSvc.CreateFromParsed(p, catID, h.defaultWalletID(), rt); err == nil {
				saved++
			}
		}
		h.send(chatID, fmt.Sprintf("✅ Saved %d/%d receipt items.", saved, len(items)))

	case data == "receipt_one_by_one":
		if h.pending == nil || len(h.pending.Items) == 0 {
			h.send(chatID, "No pending receipt items.")
			return
		}
		h.pending.OneByOne = true
		h.pending.ItemIndex = 0
		h.showNextItem(chatID)

	case data == "receipt_skip":
		if h.pending == nil || !h.pending.OneByOne {
			return
		}
		h.pending.ItemIndex++
		if h.pending.ItemIndex >= len(h.pending.Items) {
			h.pending = nil
			h.send(chatID, "✅ Done processing receipt items.")
		} else {
			h.showNextItem(chatID)
		}

	case strings.HasPrefix(data, "del:"):
		id := strings.TrimPrefix(data, "del:")
		var deleted *domain.ExpenseWithCategory
		for _, e := range h.deleteMatches {
			if e.ID == id {
				deleted = e
				break
			}
		}
		if err := h.expenseSvc.Delete(id); err != nil {
			h.send(chatID, "❌ Failed to delete. Please try again.")
			return
		}
		remaining := h.deleteMatches[:0]
		for _, e := range h.deleteMatches {
			if e.ID != id {
				remaining = append(remaining, e)
			}
		}
		h.deleteMatches = remaining
		if deleted != nil {
			desc := ""
			if deleted.Description != nil {
				desc = *deleted.Description
			}
			h.send(chatID, fmt.Sprintf("🗑 Deleted: <b>%s</b> (%.2f)",
				html.EscapeString(desc), deleted.Amount))
		} else {
			h.send(chatID, "🗑 Deleted.")
		}

	case data == "delete_done":
		h.deleteMatches = nil
		h.send(chatID, "✅ Done.")

	case data == "summary":
		h.sendSummary(chatID)

	case data == "export_csv":
		h.send(chatID, "📥 Download your CSV from the web dashboard: <i>Dashboard → Transactions → Export</i>")

	case strings.HasPrefix(data, "budget_set:"):
		catIDOrOverall := strings.TrimPrefix(data, "budget_set:")
		var catID *string
		label := "Overall"
		if catIDOrOverall != "overall" {
			catID = &catIDOrOverall
			if cat, _ := h.categorySvc.FindByID(catIDOrOverall); cat != nil {
				label = cat.Icon + " " + cat.Name
			}
		}
		h.budgetPrompt = &budgetPrompt{CategoryID: catID, Label: label}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("💬 Enter new budget amount for <b>%s</b> (in %s):", html.EscapeString(label), h.userCurrency()))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		h.bot.Send(msg) //nolint:errcheck
	}
}

func (h *Handler) handleBudgetAmountReply(msg *tgbotapi.Message) {
	p := h.budgetPrompt
	h.budgetPrompt = nil

	amount, err := strconv.ParseFloat(strings.TrimSpace(msg.Text), 64)
	if err != nil || amount <= 0 {
		h.send(msg.Chat.ID, "Invalid amount. Please enter a positive number.")
		return
	}
	now := time.Now()
	if _, err = h.budgetSvc.Upsert(p.CategoryID, amount, now.Year(), int(now.Month()), false); err != nil {
		log.Error().Err(err).Msg("bot: budget upsert failed")
		h.send(msg.Chat.ID, "Failed to save budget. Please try again.")
		return
	}
	h.send(msg.Chat.ID, fmt.Sprintf("✅ Budget for <b>%s</b> set to <b>%.2f %s</b>.", html.EscapeString(p.Label), amount, h.userCurrency()))
}

func (h *Handler) sendSummary(chatID int64) {
	now := time.Now()
	expenses, err := h.expenseSvc.MonthlySummary(now.Year(), int(now.Month()))
	if err != nil {
		h.send(chatID, "Couldn't load summary. Please try again.")
		return
	}
	incomeTotal, _ := h.walletSvc.MonthlyIncomeTotal(now.Year(), int(now.Month()))
	net := incomeTotal - expenses.Total
	text := fmt.Sprintf("📊 <b>%s Summary</b>\n\n🟢 Income: <b>%.2f</b>\n🔴 Expenses: <b>%.2f</b>\n⚖️ Net: <b>%.2f</b>\n",
		html.EscapeString(now.Format("January 2006")), incomeTotal, expenses.Total, net)
	for _, c := range expenses.Categories {
		text += fmt.Sprintf("%s <b>%s</b>: %.2f (%d)\n", c.Icon, html.EscapeString(c.CategoryName), c.Total, c.Count)
	}
	h.send(chatID, text)
}

func (h *Handler) sendBudgetStatus(chatID int64) {
	now := time.Now()
	statuses, err := h.budgetSvc.StatusForMonth(now.Year(), int(now.Month()))
	if err != nil || len(statuses) == 0 {
		h.send(chatID, "No budgets set for this month.")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "💰 <b>Budget Status — %s</b>\n\n", html.EscapeString(now.Format("January 2006")))

	var rows []tgbotapi.InlineKeyboardButton
	for _, s := range statuses {
		label, icon := "Overall", "📦"
		if s.CategoryName != "" {
			label, icon = s.CategoryName, s.CategoryIcon
		}
		emoji := "✅"
		if s.Pct >= 100 {
			emoji = "🚨"
		} else if s.Pct >= 80 {
			emoji = "⚠️"
		}
		fmt.Fprintf(&sb, "%s %s <b>%s</b>\n%s %.1f%% (%.0f / %.0f %s)\n\n",
			emoji, icon, html.EscapeString(label), progressBar(s.Pct), s.Pct, s.Spent, s.Effective, h.userCurrency())
		catKey := "overall"
		if s.CategoryID != nil {
			catKey = *s.CategoryID
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("✏️ "+label, "budget_set:"+catKey))
	}

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(rows); i += 2 {
		if i+1 < len(rows) {
			keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(rows[i], rows[i+1]))
		} else {
			keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(rows[i]))
		}
	}
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📊 View Summary", "summary"),
	))

	out := tgbotapi.NewMessage(chatID, sb.String())
	out.ParseMode = "HTML"
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
	h.bot.Send(out) //nolint:errcheck
}

func (h *Handler) sendMonthlyReport(chatID int64) {
	now := time.Now()
	expenses, err := h.expenseSvc.MonthlySummary(now.Year(), int(now.Month()))
	if err != nil {
		h.send(chatID, "Couldn't load report. Please try again.")
		return
	}
	incomeTotal, _ := h.walletSvc.MonthlyIncomeTotal(now.Year(), int(now.Month()))
	budgets, _ := h.budgetSvc.StatusForMonth(now.Year(), int(now.Month()))

	text := fmt.Sprintf("📊 <b>Report — %s</b>\n\n🟢 Income: <b>%.2f</b>\n🔴 Expenses: <b>%.2f</b>\n⚖️ Net: <b>%.2f</b>\n",
		html.EscapeString(now.Format("January 2006")), incomeTotal, expenses.Total, incomeTotal-expenses.Total)
	for _, c := range expenses.Categories {
		text += fmt.Sprintf("%s <b>%s</b>: %.2f (%d)\n", c.Icon, html.EscapeString(c.CategoryName), c.Total, c.Count)
	}
	for _, s := range budgets {
		label := "Overall"
		if s.CategoryName != "" {
			label = s.CategoryName
		}
		text += fmt.Sprintf("• %s: %.1f%%\n", html.EscapeString(label), s.Pct)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 View Summary", "summary"),
			tgbotapi.NewInlineKeyboardButtonData("📥 Export CSV", "export_csv"),
		),
	)
	out := tgbotapi.NewMessage(chatID, text)
	out.ParseMode = "HTML"
	out.ReplyMarkup = keyboard
	h.bot.Send(out) //nolint:errcheck
}

func (h *Handler) sendCategories(chatID int64) {
	cats, err := h.categorySvc.FindAll()
	if err != nil || len(cats) == 0 {
		h.send(chatID, "No categories found.")
		return
	}
	text := "🏷️ <b>Categories</b>\n\n"
	for _, c := range cats {
		text += fmt.Sprintf("%s <b>%s</b>\n", c.Icon, html.EscapeString(c.Name))
	}
	h.send(chatID, text)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func transactionType(value string) string {
	if strings.ToLower(strings.TrimSpace(value)) == "income" {
		return "income"
	}
	return "expense"
}

func transactionLabel(value string) string {
	if transactionType(value) == "income" {
		return "income"
	}
	return "expense"
}

func transactionTitle(value string) string {
	if transactionType(value) == "income" {
		return "Income"
	}
	return "Expense"
}

func transactionPlural(value string) string {
	if transactionType(value) == "income" {
		return "Income"
	}
	return "Expenses"
}

func transactionIcon(value string) string {
	if transactionType(value) == "income" {
		return "💵"
	}
	return "🧾"
}

func progressBar(pct float64) string {
	filled := int(pct / 10)
	if filled > 10 {
		filled = 10
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
}
