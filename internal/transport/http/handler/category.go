package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/aididalam/llmexpensetracker/internal/service"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/middleware"
	"github.com/aididalam/llmexpensetracker/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	svc            *service.CategoryService
	subcategorySvc *service.SubcategoryService
}

func NewCategoryHandler(svc *service.CategoryService, subcategorySvc *service.SubcategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc, subcategorySvc: subcategorySvc}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	cats, err := svc.FindAll()
	if err != nil {
		response.ResError(w, "failed to fetch categories", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, cats)
}

func (h *CategoryHandler) ListSubcategories(w http.ResponseWriter, r *http.Request) {
	if h.subcategorySvc == nil {
		response.ResSuccess(w, []*domain.Subcategory{})
		return
	}

	userID := middleware.CurrentUserID(r.Context())
	subcatSvc := h.subcategorySvc.ForUser(userID)

	id := chi.URLParam(r, "id")
	subcategories, err := subcatSvc.FindByCategory(id)
	if err != nil {
		response.ResError(w, "failed to fetch subcategories", http.StatusInternalServerError)
		return
	}
	if subcategories == nil {
		subcategories = []*domain.Subcategory{}
	}
	response.ResSuccess(w, subcategories)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	id := chi.URLParam(r, "id")

	var body struct {
		Name  string `json:"name"`
		Icon  string `json:"icon"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ResError(w, "invalid request body")
		return
	}
	if body.Name == "" {
		response.ResError(w, map[string]string{"name": "name is required"})
		return
	}

	cat, err := svc.Update(id, body.Name, body.Icon, body.Color)
	if err != nil || cat == nil {
		response.ResError(w, "category not found", http.StatusNotFound)
		return
	}
	response.ResSuccess(w, cat)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.CurrentUserID(r.Context())
	svc := h.svc.ForUser(userID)

	id := chi.URLParam(r, "id")
	if err := svc.Delete(id); err != nil {
		response.ResError(w, "failed to delete category", http.StatusInternalServerError)
		return
	}
	response.ResSuccess(w, "category deleted")
}
