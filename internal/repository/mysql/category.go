package mysql

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/jmoiron/sqlx"
)

type categoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll(userID string) ([]*domain.Category, error) {
	var cats []*domain.Category
	err := r.db.Select(&cats, "SELECT * FROM categories WHERE user_id = ? AND is_deleted = false ORDER BY name ASC", userID)
	return cats, err
}

func (r *categoryRepository) FindByName(userID, name string) (*domain.Category, error) {
	var cat domain.Category
	err := r.db.Get(&cat,
		"SELECT * FROM categories WHERE user_id = ? AND LOWER(name) = ? AND is_deleted = false LIMIT 1",
		userID, strings.ToLower(name),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &cat, err
}

func (r *categoryRepository) FindByID(userID, id string) (*domain.Category, error) {
	var cat domain.Category
	err := r.db.Get(&cat, "SELECT * FROM categories WHERE user_id = ? AND category_id = ? AND is_deleted = false LIMIT 1", userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &cat, err
}

func (r *categoryRepository) Create(cat *domain.Category) error {
	_, err := r.db.NamedExec(`
		INSERT INTO categories (category_id, user_id, name, icon, color)
		VALUES (:category_id, :user_id, :name, :icon, :color)
	`, cat)
	return err
}

func (r *categoryRepository) Update(cat *domain.Category) error {
	_, err := r.db.NamedExec(`
		UPDATE categories
		SET name = :name, icon = :icon, color = :color
		WHERE user_id = :user_id AND category_id = :category_id AND is_deleted = false
	`, cat)
	return err
}

func (r *categoryRepository) SoftDelete(userID, id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		"UPDATE categories SET is_deleted = true, deleted_at = ? WHERE user_id = ? AND category_id = ?",
		now, userID, id,
	)
	return err
}
