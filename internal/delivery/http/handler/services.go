package handler

import (
	"context"
	"time"

	"pocketly/internal/domain"
)

/*
AuthService is what AuthHandler needs from the business layer. It is
declared here, on the consumer side, so the handler depends on the
behavior it uses rather than on the concrete usecase.AuthUsecase,
which satisfies it. Tests can pass a small fake instead.
*/
type AuthService interface {
	Register(ctx context.Context, name, email, password string) (string, *domain.User, error)
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
}

/*
TransactionService is what TransactionHandler needs from the business
layer, declared on the consumer side for the same reason as
AuthService. usecase.TransactionUsecase satisfies it.
*/
type TransactionService interface {
	CreateTransaction(ctx context.Context, userID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error)
	CreateTransactionFromImage(ctx context.Context, userID string, imageData []byte) (*domain.Transaction, error)
	GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error)
	ListMyTransactions(ctx context.Context, userID string) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, userID string, transactionID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, transactionID string) error
}
