package domain

import "time"

type Category struct {
	ID        string     `db:"category_id" json:"category_id"`
	UserID    string     `db:"user_id"     json:"user_id"`
	Name      string     `db:"name"        json:"name"`
	Icon      string     `db:"icon"        json:"icon"`
	Color     string     `db:"color"       json:"color"`
	IsDeleted bool       `db:"is_deleted"  json:"is_deleted"`
	DeletedAt *time.Time `db:"deleted_at"  json:"deleted_at"`
	CreatedAt time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"  json:"updated_at"`
}

type CategoryRepository interface {
	FindAll(userID string) ([]*Category, error)
	FindByName(userID, name string) (*Category, error)
	FindByID(userID, id string) (*Category, error)
	Create(cat *Category) error
	Update(cat *Category) error
	SoftDelete(userID, id string) error
}
