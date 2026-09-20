package dto

import "time"

type TransactionItemRequest struct {
	Name     string  `json:"name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,min=1"`
	Price    float64 `json:"price" binding:"required,min=0"`
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
