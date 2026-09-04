package usecase

import (
	"context"
	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
	categorizer     repository.CategorizerRepository
}

func NewTransactionUsecase(transactionRepo repository.TransactionRepository, categorizer repository.CategorizerRepository) *TransactionUsecase {
	return &TransactionUsecase{transactionRepo: transactionRepo, categorizer: categorizer}
}

func (U *TransactionUsecase) CreateTransaction(ctx context.Context, userID string, descriptions string, items []domain.TransactionItem) (*domain.Transaction, error) {
	panic("not implemented")
}

func (U *TransactionUsecase) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	panic("not implemented")
}

func (U *TransactionUsecase) ListMyTransaction(ctx context.Context, userID string) (*domain.Transaction, error) {
	panic("not implemented")
}

func (U *TransactionUsecase) UpdateTransaction(ctx context.Context, userID string, descriptions string, items []domain.TransactionItem) (*domain.Transaction, error) {
	panic("not implemented")
}

func (U *TransactionUsecase) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	panic("not implemented")
}
