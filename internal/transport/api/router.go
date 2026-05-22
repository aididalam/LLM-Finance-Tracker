package api

import (
	"net/http"
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func Router(cfg *config.Config, db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)

	files := http.Dir("internal/web/files")

	r.Handle("/*", ServeFrontend(files))

	return r
}

func ServeFrontend(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}

		if f, err := root.Open(name); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, "internal/web/files/index.html")
	})
}
