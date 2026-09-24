package persistence

import (
	"time"
)

/*
UserModel is the GORM row representation of a user account.
It intentionally stays separate from domain.User: this struct carries
database-specific tags (primary key, unique index) that the business
layer must never depend on. A user owns many Transactions; each
Transaction belongs to exactly one user via UserID. Transactions is
only declared so GORM creates the foreign key — it is never preloaded.
Deleting a user cascades to their Transactions at the database level.
*/
type UserModel struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string
	CreatedAt time.Time

	Transactions []TransactionModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

/*
TableName overrides GORM's default pluralization behavior,
mapping UserModel explicitly to the "users" table.
*/
func (UserModel) TableName() string {
	return "users"
}

/*
TransactionModel is the GORM row representation of a financial
transaction. TotalAmount is derived from the sum of its related
Items — it is not meant to be set independently. Items is preloaded
via GORM's relation feature; callers must use Preload("Items") when
querying, since it is not loaded automatically. Deleting a Transaction
cascades to its Items at the database level.
*/
type TransactionModel struct {
	ID          string `gorm:"primaryKey"`
	UserID      string `gorm:"index"`
	Description string
	Category    string
	TotalAmount float64
	Date        time.Time
	CreatedAt   time.Time

	Items []TransactionItemModel `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

/*
TableName overrides GORM's default pluralization behavior,
mapping TransactionModel explicitly to the "transactions" table.
*/
func (TransactionModel) TableName() string {
	return "transactions"
}

/*
TransactionItemModel is the GORM row representation of a single line
item belonging to a Transaction (e.g. one product from a receipt).
It always belongs to exactly one Transaction via TransactionID.
*/
type TransactionItemModel struct {
	ID            string `gorm:"primaryKey"`
	TransactionID string `gorm:"index"`
	Name          string
	Quantity      int
	Price         float64
}

/*
TableName overrides GORM's default pluralization behavior,
mapping TransactionItemModel explicitly to the "transaction_items" table.
*/
func (TransactionItemModel) TableName() string {
	return "transaction_items"
}
