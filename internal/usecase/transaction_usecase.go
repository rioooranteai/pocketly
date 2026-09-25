package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"pocketly/internal/domain"
	"pocketly/internal/port"
	"pocketly/internal/repository"
)

/*
TransactionUsecase implements the business logic for creating, reading,
updating, and deleting financial transactions. It enforces ownership
checks so a user can never access another user's transactions, and
delegates automatic category detection to a Categorizer implementation
and receipt image parsing to a VisionExtractor implementation.
*/
type TransactionUsecase struct {
	transactionRepo repository.TransactionRepository
	categorizer     port.Categorizer
	visionExtractor port.VisionExtractor
	maxImageSize    int64
}

/*
NewTransactionUsecase builds a TransactionUsecase backed by the given
repository, categorizer, and vision extractor. maxImageSize is the
largest receipt image, in bytes, that CreateTransactionFromImage
accepts.
*/
func NewTransactionUsecase(transactionRepo repository.TransactionRepository, categorizer port.Categorizer, visionExtractor port.VisionExtractor, maxImageSize int64) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo: transactionRepo,
		categorizer:     categorizer,
		visionExtractor: visionExtractor,
		maxImageSize:    maxImageSize,
	}
}

/*
resolveCategory determines the category for a transaction description
via the Categorizer. If the Categorizer fails for any reason, it logs
the error and falls back to domain.CategoryUncategorized instead of
propagating the failure — categorization is a best-effort enrichment,
not a precondition for saving a transaction. A category outside the
domain list is treated the same way, so a misbehaving implementation
can never store an unknown category.
*/
func (uc *TransactionUsecase) resolveCategory(ctx context.Context, description string) string {
	category, err := uc.categorizer.Categorize(ctx, description)
	if err != nil {
		log.Printf("categorizer failed, falling back to %q: %v", domain.CategoryUncategorized, err)
		return domain.CategoryUncategorized
	}
	if !domain.IsValidCategory(category) {
		log.Printf("categorizer returned unknown category %q, falling back to %q", category, domain.CategoryUncategorized)
		return domain.CategoryUncategorized
	}
	return category
}

/*
assignItemIDs gives every item a fresh ID and links it to the given
transaction. Items arrive without IDs (handlers never set one), and
an empty string primary key makes every item after the first collide
on insert, which GORM silently skips. IDs are generated here for the
same reason transaction and user IDs are: identity is an application
decision, not something a delivery or AI adapter should own.
*/
func assignItemIDs(transaction *domain.Transaction) {
	for i := range transaction.Items {
		transaction.Items[i].ID = uuid.New().String()
		transaction.Items[i].TransactionID = transaction.ID
	}
}

/*
validateItems checks every item against the domain's item rules
before anything is saved. It runs in the usecase rather than relying
on DTO binding tags, because items reach this layer from more than
one source: the HTTP body, and the vision extractor on /scan, whose
output never passes through a DTO. The error names the 1-based item
position so the caller can tell which one was rejected.
*/
func validateItems(items []domain.TransactionItem) error {
	for i, item := range items {
		if !item.ValidateItemData() {
			return fmt.Errorf("item %d: %w", i+1, domain.ErrInvalidItemData)
		}
	}
	return nil
}

/*
CreateTransaction records a new transaction for the given user. The
total amount is derived from the sum of its items rather than accepted
directly, and the category is determined automatically via the
Categorizer based on the transaction's description. Categorization
failures never block the transaction from being saved.
*/
func (uc *TransactionUsecase) CreateTransaction(ctx context.Context, userID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error) {
	if err := validateItems(items); err != nil {
		return nil, err
	}

	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: description,
		Date:        date,
		Items:       items,
		CreatedAt:   time.Now().UTC(),
		Category:    uc.resolveCategory(ctx, description),
	}

	assignItemIDs(transaction)
	transaction.CalculateTotal()

	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

/*
CreateTransactionFromImage records a new transaction whose description
and items are extracted from a receipt image via VisionExtractor,
rather than typed in manually. The image is checked for being empty
or larger than maxImageSize before any provider is called, so the
rule holds whichever VisionExtractor is wired in. Once extracted, it
follows the same persistence path as CreateTransaction: total is
recalculated from the extracted items, and category is resolved the
same way, with the same fallback behavior on failure.
*/
func (uc *TransactionUsecase) CreateTransactionFromImage(ctx context.Context, userID string, imageData []byte, date time.Time) (*domain.Transaction, error) {
	if len(imageData) == 0 {
		return nil, domain.ErrEmptyImageData
	}
	if int64(len(imageData)) > uc.maxImageSize {
		return nil, domain.ErrImageSizeExceedsLimit
	}

	description, items, err := uc.visionExtractor.Extract(ctx, imageData)
	if err != nil {
		return nil, err
	}
	if err := validateItems(items); err != nil {
		return nil, err
	}

	transaction := &domain.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: description,
		Date:        date,
		Items:       items,
		CreatedAt:   time.Now().UTC(),
		Category:    uc.resolveCategory(ctx, description),
	}

	assignItemIDs(transaction)
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
domain.CategoryUncategorized instead of blocking the update.
*/
func (uc *TransactionUsecase) UpdateTransaction(ctx context.Context, userID string, transactionID string, description string, items []domain.TransactionItem, date time.Time) (*domain.Transaction, error) {
	transaction, err := uc.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}
	if err := validateItems(items); err != nil {
		return nil, err
	}

	transaction.Description = description
	transaction.Date = date
	transaction.Items = items
	transaction.Category = uc.resolveCategory(ctx, description)
	assignItemIDs(transaction)
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
