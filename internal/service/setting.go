package service

import (
	"strconv"

	"github.com/aididalam/llmexpensetracker/internal/domain"
)

// SettingService provides typed access to application settings.
type SettingService struct {
	userID string
	repo   domain.SettingRepository
}

// NewSettingService creates a new SettingService.
func NewSettingService(repo domain.SettingRepository) *SettingService {
	return &SettingService{repo: repo, userID: domain.DefaultUserID}
}

// ForUser returns a scoped copy of SettingService for the given user.
func (s *SettingService) ForUser(userID string) *SettingService {
	scoped := *s
	scoped.userID = userID
	return &scoped
}

// GetAll returns all settings as a raw map.
func (s *SettingService) GetAll() (map[string]string, error) {
	return s.repo.GetAll(s.userID)
}

// Set persists a single setting value.
func (s *SettingService) Set(key, value string) error {
	return s.repo.Set(s.userID, key, value)
}

// GetString returns the value for a key, or defaultVal if not found.
func (s *SettingService) GetString(key, defaultVal string) string {
	m, err := s.repo.GetAll(s.userID)
	if err != nil || m == nil {
		return defaultVal
	}
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return defaultVal
}

// GetFloat returns the float64 value for a key, or defaultVal if not found/invalid.
func (s *SettingService) GetFloat(key string, defaultVal float64) float64 {
	v := s.GetString(key, "")
	if v == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

// GetBool returns the bool value for a key, or defaultVal if not found/invalid.
func (s *SettingService) GetBool(key string, defaultVal bool) bool {
	v := s.GetString(key, "")
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}
