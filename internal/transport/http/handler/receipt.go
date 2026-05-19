package handler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/llm"
	"github.com/aididalam/llmexpensetracker/internal/repository/mysql"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
)

const maxReceiptSize = 10 << 20 // 10 MB

type ReceiptHandler struct {
	llmProvider llm.Provider
	categorySvc *service.CategoryService
	settingSvc  *service.SettingService
	llmUsage    *mysql.LLMUsageRepository
}

func NewReceiptHandler(llmProvider llm.Provider, categorySvc *service.CategoryService, llmUsage *mysql.LLMUsageRepository, settingSvc ...*service.SettingService) *ReceiptHandler {
	h := &ReceiptHandler{llmProvider: llmProvider, categorySvc: categorySvc, llmUsage: llmUsage}
	if len(settingSvc) > 0 {
		h.settingSvc = settingSvc[0]
	}
	return h
}

// Parse reads an uploaded receipt image and returns parsed expense details.
//
//	POST /api/v1/receipt/parse   multipart/form-data, field: "receipt"
func (h *ReceiptHandler) Parse(w http.ResponseWriter, r *http.Request) {
	if h.llmProvider == nil {
		response.ResError(w, "LLM not configured", http.StatusServiceUnavailable)
		return
	}
	userID := middleware.CurrentUserID(r.Context())

	currency := "USD"
	if h.settingSvc != nil {
		currency = h.settingSvc.ForUser(userID).GetString("currency", "USD")
	}

	if err := r.ParseMultipartForm(maxReceiptSize); err != nil {
		response.ResError(w, "file too large (max 10 MB)")
		return
	}

	file, header, err := r.FormFile("receipt")
	if err != nil {
		response.ResError(w, "receipt file is required")
		return
	}
	defer file.Close()

	mediaType := detectMediaType(header.Header.Get("Content-Type"), header.Filename)
	if mediaType == "" {
		response.ResError(w, "unsupported file type — upload a JPEG, PNG, WEBP image or PDF")
		return
	}
	receiptType := "image"
	if mediaType == "application/pdf" {
		receiptType = "pdf"
	}

	data, err := io.ReadAll(io.LimitReader(file, maxReceiptSize))
	if err != nil {
		response.ResError(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	categories, err := h.categorySvc.Names()
	if err != nil {
		response.ResError(w, "failed to load categories", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()

	parsed, usage, err := h.llmProvider.ParseReceipt(ctx, data, mediaType, categories, currency)
	if usage != nil && h.llmUsage != nil {
		_ = h.llmUsage.ForUser(userID).Log(h.llmProvider.Name(), h.llmProvider.Model(), usage.PromptTokens, usage.OutputTokens, "web_receipt", receiptType)
	}
	if err != nil {
		// PreprocessReceipt returns user-friendly errors (e.g. PDF too large).
		// Distinguish between client errors (4xx) and server errors (5xx).
		msg := err.Error()
		status := http.StatusInternalServerError
		if len(msg) > 0 && (strings.Contains(msg, "too large") || strings.Contains(msg, "unsupported")) {
			status = http.StatusRequestEntityTooLarge
		}
		response.ResError(w, msg, status)
		return
	}

	response.ResSuccess(w, parsed)
}

func detectMediaType(contentType, filename string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return "image/jpeg"
	case strings.Contains(ct, "png"):
		return "image/png"
	case strings.Contains(ct, "gif"):
		return "image/gif"
	case strings.Contains(ct, "webp"):
		return "image/webp"
	case strings.Contains(ct, "pdf"):
		return "application/pdf"
	}
	// fall back to extension
	fn := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(fn, ".jpg") || strings.HasSuffix(fn, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(fn, ".png"):
		return "image/png"
	case strings.HasSuffix(fn, ".gif"):
		return "image/gif"
	case strings.HasSuffix(fn, ".webp"):
		return "image/webp"
	case strings.HasSuffix(fn, ".pdf"):
		return "application/pdf"
	}
	return ""
}
