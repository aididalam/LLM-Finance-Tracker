package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type pending struct {
	Parsed                *llm.ParsedExpense
	CategoryID            *string
	ReceiptType           string
	Items                 []llm.ReceiptItem
	ItemIndex             int
	OneByOne              bool
	WalletID              string
	WalletBankDebitCardID *string
}

type budgetPrompt struct {
	CategoryID *string
	Label      string
}

type Handler struct {
	bot                 *tgbotapi.BotAPI
	chatID              int64
	llmProvider         llm.Provider
	categorySvc         *service.CategoryService
	expenseSvc          *service.ExpenseService
	budgetSvc           *service.BudgetService
	walletSvc           *service.WalletService
	settingSvc          *service.SettingService
	llmUsage            *mysql.LLMUsageRepository
	pending             *pending
	budgetPrompt        *budgetPrompt
	conversationHistory []llm.ChatMessage
	deleteMatches       []*domain.ExpenseWithCategory
	mediaGroupBuf       map[string][]string
}

func NewHandler(
	cfg *config.Config,
	llmProvider llm.Provider,
	categorySvc *service.CategoryService,
	expenseSvc *service.ExpenseService,
	budgetSvc *service.BudgetService,
	walletSvc *service.WalletService,
	llmUsage *mysql.LLMUsageRepository,
	settingSvc ...*service.SettingService,
) (*Handler, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("initialize telegram bot: %w", err)
	}
	chatID, err := strconv.ParseInt(cfg.TelegramChatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse TELEGRAM_CHAT_ID: %w", err)
	}
	h := &Handler{
		bot:         bot,
		chatID:      chatID,
		llmProvider: llmProvider,
		categorySvc: categorySvc,
		expenseSvc:  expenseSvc,
		budgetSvc:   budgetSvc,
		walletSvc:   walletSvc,
		llmUsage:    llmUsage,
	}
	if len(settingSvc) > 0 {
		h.settingSvc = settingSvc[0]
	}
	log.Info().Msgf("telegram bot authorized as @%s", bot.Self.UserName)
	return h, nil
}

func (h *Handler) BotAPI() *tgbotapi.BotAPI { return h.bot }
func (h *Handler) ChatID() int64            { return h.chatID }

func (h *Handler) userCurrency() string {
	if h.settingSvc == nil {
		return "USD"
	}
	return h.settingSvc.GetString("currency", "USD")
}

func (h *Handler) defaultWalletID() string {
	id, err := h.walletSvc.DefaultWalletID()
	if err != nil || id == "" {
		return ""
	}
	return id
}

func (h *Handler) newChatProcessor() *service.ChatProcessor {
	return service.NewChatProcessor(h.llmProvider, h.categorySvc, h.expenseSvc, h.walletSvc)
}

func (h *Handler) addToHistory(role, content string) {
	h.conversationHistory = append(h.conversationHistory, llm.ChatMessage{Role: role, Content: content})
	if len(h.conversationHistory) > 6 {
		h.conversationHistory = h.conversationHistory[len(h.conversationHistory)-6:]
	}
}

func (h *Handler) clearHistory() {
	h.conversationHistory = nil
}

func (h *Handler) StartPolling(ctx context.Context) {
	log.Info().Msg("bot: polling started")
	var offset int
	backoff := time.Second

	for {
		if ctx.Err() != nil {
			break
		}
		updates, err := h.bot.GetUpdates(tgbotapi.UpdateConfig{Offset: offset, Timeout: 60})
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Warn().Err(err).Dur("retry_in", backoff).Msg("bot: getUpdates failed, retrying")
			select {
			case <-ctx.Done():
				goto done
			case <-time.After(backoff):
			}
			if backoff < 64*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
		for _, update := range updates {
			h.processUpdate(update)
			offset = update.UpdateID + 1
		}
		h.flushMediaGroups()
	}

done:
	log.Info().Msg("bot: polling stopped")
}

func (h *Handler) processUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.Chat.ID != h.chatID {
			log.Warn().Int64("chat_id", update.Message.Chat.ID).Msg("bot: unauthorized chat")
			return
		}
		h.handleMessage(update.Message)
	}
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Chat.ID != h.chatID {
			return
		}
		h.handleCallback(update.CallbackQuery)
	}
}
