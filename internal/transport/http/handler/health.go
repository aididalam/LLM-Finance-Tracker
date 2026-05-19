package handler

import (
	"net/http"

	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response.ResSuccess(w, "ok")
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		response.ResError(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	response.ResSuccess(w, "ok")
}
