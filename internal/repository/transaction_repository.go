package repository

import (
	"context"

	"pocketly/internal/domain"
)

/*
TransactionRepository is the contract for persisting and retrieving
financial transactions. The usecase layer depends only on this
interface, never on a concrete implementation, so the underlying
storage can be swapped without touching business logic.
*/
type TransactionRepository interface {
	/*
		Create persists a new transaction record, along with its items.
	*/
	Create(ctx context.Context, t *domain.Transaction) error

	/*
		FindByID looks up a transaction by its unique identifier.
		Implementations should return (nil, nil) when no transaction
		matches, distinguishing "not found" from an actual error.
	*/
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)

	/*
		ListByUser returns all transactions belonging to the given user.
	*/
	ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error)

	/*
		Update persists changes to an existing transaction, including
		its items.
	*/
	Update(ctx context.Context, t *domain.Transaction) error

	/*
		Delete removes a transaction and its related items.
	*/
	Delete(ctx context.Context, id string) error
}
