package domain

type TransactionItem struct {
	ID            string
	TransactionID string
	Name          string
	Quantity      int
	Price         float64
}

func (ti TransactionItem) Subtotal() float64 {
	return ti.Price * float64(ti.Quantity)
}
