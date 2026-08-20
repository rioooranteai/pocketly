package persistence

import (
	"time"
)

/*
UserModel is the GORM row representation of a user account.
It intentionally stays separate from domain.User: this struct carries
database-specific tags (primary key, unique index) that the business
layer must never depend on.
*/
type UserModel struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	CreatedAt time.Time
}

/*
TableName overrides GORM's default pluralization behavior,
mapping UserModel explicitly to the "users" table.
*/
func (UserModel) TableName() string {
	return "users"
}
