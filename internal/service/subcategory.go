package service

import (
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/google/uuid"
)

type SubcategoryService struct {
	userID string
	repo   domain.SubcategoryRepository
}

func NewSubcategoryService(repo domain.SubcategoryRepository) *SubcategoryService {
	return &SubcategoryService{repo: repo, userID: domain.DefaultUserID}
}

func (s *SubcategoryService) ForUser(userID string) *SubcategoryService {
	scoped := *s
	scoped.userID = userID
	return &scoped
}

func (s *SubcategoryService) FindOrCreate(categoryID *string, name string) (*domain.Subcategory, error) {
	name = cleanSubcategoryName(name)
	if categoryID == nil || *categoryID == "" || name == "" {
		return nil, nil
	}

	subcat, err := s.repo.FindByName(s.userID, *categoryID, name)
	if err != nil {
		return nil, err
	}
	if subcat != nil {
		return subcat, nil
	}

	subcat = &domain.Subcategory{
		ID:         uuid.New().String(),
		UserID:     s.userID,
		CategoryID: *categoryID,
		Name:       name,
	}
	if err := s.repo.Create(subcat); err != nil {
		return nil, err
	}
	return subcat, nil
}

func (s *SubcategoryService) FindByCategory(categoryID string) ([]*domain.Subcategory, error) {
	return s.repo.FindByCategory(s.userID, categoryID)
}

func (s *SubcategoryService) FindByID(id string) (*domain.Subcategory, error) {
	return s.repo.FindByID(s.userID, id)
}

func cleanSubcategoryName(name string) string {
	name = strings.TrimSpace(name)
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = strings.TrimSpace(name[idx+1:])
	}
	if idx := strings.Index(name, "›"); idx >= 0 {
		name = strings.TrimSpace(name[idx+len("›"):])
	}
	return name
}
