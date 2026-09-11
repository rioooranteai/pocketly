package domain

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
ValidateItemData reports whether the item's quantity and price are
non-negative. It does not check for zero values, since a free item
(price 0) or a placeholder entry (quantity 0) may still be valid.
*/
func (ti TransactionItem) ValidateItemData() bool {
	if ti.Quantity < 0 || ti.Price < 0 {
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
