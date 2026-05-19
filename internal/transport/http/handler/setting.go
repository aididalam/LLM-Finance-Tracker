package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
)

// SettingHandler handles settings CRUD over HTTP.
type SettingHandler struct {
	svc *service.SettingService
}

// NewSettingHandler creates a new SettingHandler.
func NewSettingHandler(svc *service.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

// allowed keys that the frontend may update.
var allowedSettingKeys = map[string]bool{
	"currency":               true,
	"budget_alert_threshold": true,
	"budget_alert_enabled":   true,
	"timezone":               true,
	"locale":                 true,
	"telegram_chat_id":       true,
	"metadata":               true,
}

// GetAll godoc  GET /api/v1/settings
func (h *SettingHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	settings, err := svc.GetAll()
	if err != nil {
		response.ResError(w, "failed to fetch settings", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, settings)
}

// Update godoc  PUT /api/v1/settings
func (h *SettingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body")
		return
	}
	for k, v := range body {
		if !allowedSettingKeys[k] {
			continue
		}
		value, err := settingValueString(v)
		if err != nil {
			response.ResError(w, "invalid setting value")
			return
		}
		if err := svc.Set(k, value); err != nil {
			response.ResError(w, "failed to save setting", http.StatusInternalServerError)
			return
		}
	}
	response.ResSuccess(w, "ok")
}

func settingValueString(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
