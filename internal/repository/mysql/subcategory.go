package mysql

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/aididalam/llmexpensetracker/internal/domain"
	"github.com/jmoiron/sqlx"
)

type subcategoryRepository struct {
	db *sqlx.DB
}

func NewSubcategoryRepository(db *sqlx.DB) domain.SubcategoryRepository {
	return &subcategoryRepository{db: db}
}

func (r *subcategoryRepository) FindByName(userID, categoryID, name string) (*domain.Subcategory, error) {
	var subcat domain.Subcategory
	err := r.db.Get(&subcat, `
		SELECT * FROM subcategories
		WHERE user_id = ? AND category_id = ? AND LOWER(name) = ? AND is_deleted = false
		LIMIT 1
	`, userID, categoryID, strings.ToLower(name))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &subcat, err
}

func (r *subcategoryRepository) FindByID(userID, id string) (*domain.Subcategory, error) {
	var subcat domain.Subcategory
	err := r.db.Get(&subcat, "SELECT * FROM subcategories WHERE user_id = ? AND subcategory_id = ? AND is_deleted = false LIMIT 1", userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &subcat, err
}

func (r *subcategoryRepository) FindByCategory(userID, categoryID string) ([]*domain.Subcategory, error) {
	var subcats []*domain.Subcategory
	err := r.db.Select(&subcats, `
		SELECT * FROM subcategories
		WHERE user_id = ? AND category_id = ? AND is_deleted = false
		ORDER BY name ASC
	`, userID, categoryID)
	return subcats, err
}

func (r *subcategoryRepository) Create(subcat *domain.Subcategory) error {
	_, err := r.db.NamedExec(`
		INSERT INTO subcategories (subcategory_id, user_id, category_id, name)
		VALUES (:subcategory_id, :user_id, :category_id, :name)
	`, subcat)
	return err
}

func (r *subcategoryRepository) Update(subcat *domain.Subcategory) error {
	_, err := r.db.NamedExec(`
		UPDATE subcategories
		SET name = :name
		WHERE user_id = :user_id AND subcategory_id = :subcategory_id AND is_deleted = false
	`, subcat)
	return err
}

func (r *subcategoryRepository) SoftDelete(userID, id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		"UPDATE subcategories SET is_deleted = true, deleted_at = ? WHERE user_id = ? AND subcategory_id = ?",
		now, userID, id,
	)
	return err
}
