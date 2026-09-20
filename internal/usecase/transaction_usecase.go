package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

/*
uncategorizedFallback is used whenever the Categorizer fails to
produce a category (e.g. AI provider outage, timeout, rate limit).
Categorization is a secondary enrichment feature and must never
block the core act of recording a transaction.
*/
const uncategorizedFallback = "uncategorized"

/*
TransactionUsecase implements the business logic for creating, reading,
updating, and deleting financial transactions. It enforces ownership
checks so a user can never access another user's transactions, and
delegates automatic category detection to a Categorizer implementation.
*/
type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
	categorizer     repository.CategorizerRepository
}

/*
NewTransactionUsecase builds a TransactionUsecase backed by the given
repository and categorizer.
*/
func NewTransactionUsecase(transactionRepo repository.TransactionRepository, categorizer repository.CategorizerRepository) *TransactionUsecase {
	return &TransactionUsecase{transactionRepo: transactionRepo, categorizer: categorizer}
}

/*
resolveCategory determines the category for a transaction description
via the Categorizer. If the Categorizer fails for any reason, it logs
the error and falls back to uncategorizedFallback instead of
propagating the failure — categorization is a best-effort enrichment,
not a precondition for saving a transaction.
*/
func (uc *TransactionUsecase) resolveCategory(ctx context.Context, description string) string {
	category, err := uc.categorizer.Categorize(ctx, description)
	if err != nil {
		log.Printf("categorizer failed, falling back to %q: %v", uncategorizedFallback, err)
		return uncategorizedFallback
	}
	return category
}

/*
CreateTransaction records a new transaction for the given user. The
total amount is derived from the sum of its items rather than accepted
directly, and the category is determined automatically via the
Categorizer based on the transaction's description. Categorization
failures never block the transaction from being saved.
*/
func (uc *TransactionUsecase) CreateTransaction(ctx context.Context, userID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error) {
	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: description,
		Date:        date,
		Items:       items,
		CreatedAt:   time.Now().UTC(),
		Category:    uc.resolveCategory(ctx, description),
	}

	transaction.CalculateTotal()

	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

/*
GetTransaction retrieves a single transaction by ID, ensuring it
belongs to the requesting user.
*/
func (uc *TransactionUsecase) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	transaction, err := uc.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, domain.ErrTransactionNotFound
	}
	if transaction.UserID != userID {
		return nil, domain.ErrUnauthorizedAccess
	}

	return transaction, nil
}

/*
ListMyTransactions returns all transactions belonging to the given user.
*/
func (uc *TransactionUsecase) ListMyTransactions(ctx context.Context, userID string) ([]domain.Transaction, error) {
	return uc.transactionRepo.ListByUser(ctx, userID)
}

/*
UpdateTransaction modifies an existing transaction owned by the given
user. Ownership is verified via GetTransaction before any change is
applied. The total amount is recalculated from the updated items, and
the category is re-detected from the updated description. As with
CreateTransaction, a categorizer failure falls back to
uncategorizedFallback instead of blocking the update.
*/
func (uc *TransactionUsecase) UpdateTransaction(ctx context.Context, userID string, transactionID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error) {
	transaction, err := uc.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}

	transaction.Description = description
	transaction.Date = date
	transaction.Items = items
	transaction.Category = uc.resolveCategory(ctx, description)
	transaction.CalculateTotal()

	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

/*
DeleteTransaction removes a transaction owned by the given user.
Ownership is verified via GetTransaction before deletion.
*/
func (uc *TransactionUsecase) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	transaction, err := uc.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		return err
	}

	return uc.transactionRepo.Delete(ctx, transaction.ID)
}
