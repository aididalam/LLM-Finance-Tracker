package domain

import "time"

const DefaultUserID = "00000000-0000-4000-8000-000000000001"

type User struct {
	ID              string     `db:"user_id"           json:"user_id"`
	FirstName       string     `db:"first_name"        json:"first_name"`
	LastName        string     `db:"last_name"         json:"last_name"`
	DisplayName     string     `db:"display_name"      json:"display_name"`
	Email           string     `db:"email"             json:"email"`
	PasswordHash    string     `db:"password_hash"     json:"-"`
	Status          string     `db:"status"            json:"status"`
	AvatarURL       string     `db:"avatar_url"        json:"avatar_url"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `db:"last_login_at"     json:"last_login_at,omitempty"`
	IsDeleted       bool       `db:"is_deleted"        json:"is_deleted"`
	DeletedAt       *time.Time `db:"deleted_at"        json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"        json:"updated_at"`
}
