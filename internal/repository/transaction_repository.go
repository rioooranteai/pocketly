package repository

import (
	"context"
	"pocketly/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, t *domain.Transaction) error
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error)
	Update(ctx context.Context, t *domain.Transaction) error
	Delete(ctx context.Context, id string) error
}
