package domain

import "time"

type Subcategory struct {
	ID         string     `db:"subcategory_id" json:"subcategory_id"`
	UserID     string     `db:"user_id"        json:"user_id"`
	CategoryID string     `db:"category_id"    json:"category_id"`
	Name       string     `db:"name"           json:"name"`
	IsDeleted  bool       `db:"is_deleted"     json:"is_deleted"`
	DeletedAt  *time.Time `db:"deleted_at"     json:"deleted_at"`
	CreatedAt  time.Time  `db:"created_at"     json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"     json:"updated_at"`
}

type SubcategoryRepository interface {
	FindByName(userID, categoryID, name string) (*Subcategory, error)
	FindByID(userID, id string) (*Subcategory, error)
	FindByCategory(userID, categoryID string) ([]*Subcategory, error)
	Create(subcat *Subcategory) error
	Update(subcat *Subcategory) error
	SoftDelete(userID, id string) error
}
