package domain

import (
	"math"
	"strings"
)

/*
TransactionItem represents a single line item belonging to a
Transaction (e.g. one product from a receipt).
*/
type TransactionItem struct {
	ID            string
	TransactionID string
	Name          string
	Quantity      int
	Price         float64
}

/*
ValidateItemData reports whether the item has a non-blank name, a
non-negative quantity and price, and a finite price and subtotal.
It does not check for zero values, since a free item (price 0) or a
placeholder entry (quantity 0) may still be valid. The finite check
matters because a huge price times a quantity overflows to +Inf,
which cannot be stored meaningfully or encoded as JSON.
*/
func (ti TransactionItem) ValidateItemData() bool {
	if strings.TrimSpace(ti.Name) == "" {
		return false
	}
	if ti.Quantity < 0 || ti.Price < 0 {
		return false
	}
	if !isFinite(ti.Price) || !isFinite(ti.Subtotal()) {
		return false
	}

	return true
}

/*
Subtotal returns the line total for this item (price times quantity).
*/
func (ti TransactionItem) Subtotal() float64 {
	return ti.Price * float64(ti.Quantity)
}

/*
isFinite reports whether v is a real number, not +Inf, -Inf or NaN.
*/
func isFinite(v float64) bool {
	return !math.IsInf(v, 0) && !math.IsNaN(v)
}
