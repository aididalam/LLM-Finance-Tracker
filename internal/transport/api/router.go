package api

import (
	"encoding/json"
	"net/http"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func Router(cfg *config.Config, db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]string{
			"message": "hello",
		})
	})

	return r
}
