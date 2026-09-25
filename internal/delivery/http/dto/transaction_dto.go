package dto

import "time"

/*
TransactionItemRequest is one line item in a create or update body.
Quantity and Price are pointers so "required" means "present" rather
than "non-zero": a free item (price 0) or a placeholder (quantity 0)
is valid in the domain, while an omitted field is still rejected.
min=0 mirrors TransactionItem.ValidateItemData, which the usecase
also runs, so HTTP and non-HTTP paths enforce the same rule.
*/
type TransactionItemRequest struct {
	Name     string   `json:"name" binding:"required"`
	Quantity *int     `json:"quantity" binding:"required,min=0"`
	Price    *float64 `json:"price" binding:"required,min=0"`
}

type CreateTransactionRequest struct {
	Description string                   `json:"description" binding:"required"`
	Date        time.Time                `json:"date" binding:"required"`
	Items       []TransactionItemRequest `json:"items" binding:"required,min=1,dive"`
}

type TransactionItemResponse struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type TransactionResponse struct {
	ID          string                    `json:"id"`
	Description string                    `json:"description"`
	Category    string                    `json:"category"`
	TotalAmount float64                   `json:"total_amount"`
	Date        time.Time                 `json:"date"`
	Items       []TransactionItemResponse `json:"items"`
}
