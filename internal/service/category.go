package service

import (
	"strings"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/google/uuid"
)

type CategoryService struct {
	userID string
	repo   domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo, userID: domain.DefaultUserID}
}

func (s *CategoryService) ForUser(userID string) *CategoryService {
	scoped := *s
	scoped.userID = userID
	return &scoped
}

func (s *CategoryService) FindAll() ([]*domain.Category, error) {
	return s.repo.FindAll(s.userID)
}

func (s *CategoryService) Names() ([]string, error) {
	cats, err := s.repo.FindAll(s.userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(cats))
	for i, c := range cats {
		names[i] = c.Name
	}
	return names, nil
}

func (s *CategoryService) FindOrCreate(name string) (*domain.Category, error) {
	name = strings.TrimSpace(name)
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = strings.TrimSpace(name[idx+1:])
	}
	if name == "" {
		name = "Other"
	}

	cat, err := s.repo.FindByName(s.userID, name)
	if err != nil {
		return nil, err
	}
	if cat != nil {
		return cat, nil
	}

	cat = &domain.Category{
		ID:     uuid.New().String(),
		UserID: s.userID,
		Name:   name,
	}
	if err := s.repo.Create(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) FindByName(name string) (*domain.Category, error) {
	return s.repo.FindByName(s.userID, name)
}

func (s *CategoryService) FindByID(id string) (*domain.Category, error) {
	return s.repo.FindByID(s.userID, id)
}

func (s *CategoryService) Update(id, name, icon, color string) (*domain.Category, error) {
	cat, err := s.repo.FindByID(s.userID, id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, nil
	}
	cat.Name = name
	cat.UserID = s.userID
	cat.Icon = icon
	cat.Color = color
	return cat, s.repo.Update(cat)
}

func (s *CategoryService) Delete(id string) error {
	return s.repo.SoftDelete(s.userID, id)
}
