package domain

import (
	"time"
)

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

func (t *Transaction) CalculateTotal() {
	t.TotalAmount = 0

	for _, item := range t.Items {
		t.TotalAmount += item.Subtotal()
	}
}
