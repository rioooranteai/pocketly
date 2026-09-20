package domain

import (
	"time"
)

/*
Transaction represents a single financial transaction recorded by a
user. TotalAmount is not meant to be set directly — it is always
derived from the sum of its Items via CalculateTotal.
*/
type Transaction struct {
	ID          string
	UserID      string
	Description string
	Category    string
	TotalAmount float64
	Items       []TransactionItem
	Date        time.Time
	CreatedAt   time.Time
}

/*
CalculateTotal recomputes TotalAmount as the sum of each item's
subtotal. It must be called whenever Items changes, since
TotalAmount is not kept in sync automatically.
*/
func (t *Transaction) CalculateTotal() {
	t.TotalAmount = 0

	for _, item := range t.Items {
		t.TotalAmount += item.Subtotal()
	}
}
