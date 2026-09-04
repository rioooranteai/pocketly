package domain

type TransactionItem struct {
	ID            string
	TransactionID string
	Name          string
	Quantity      int
	Price         float64
}

func (ti TransactionItem) ValidateItemData() bool {
	if ti.Quantity < 0 || ti.Price < 0 {
		return false
	} 

	return true
}

func (ti TransactionItem) Subtotal() float64 {
	return ti.Price * float64(ti.Quantity)
}
