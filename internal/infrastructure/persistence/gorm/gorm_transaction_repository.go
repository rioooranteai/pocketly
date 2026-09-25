package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

type GormTransactionRepository struct {
	db *gorm.DB
}

var _ repository.TransactionRepository = (*GormTransactionRepository)(nil)

/*
NewGormTransactionRepository builds a GormTransactionRepository backed
by the given GORM connection. It satisfies the
repository.TransactionRepository interface.
*/
func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

/*
toTransactionModel maps a domain.Transaction entity into its GORM
persistence representation (TransactionModel), including its items,
so the database layer never depends directly on the business entity's
shape.
*/
func toTransactionModel(t *domain.Transaction) *TransactionModel {
	var itemsData []TransactionItemModel
	for _, item := range t.Items {
		itemsData = append(itemsData, toTransactionItemModel(&item))
	}

	return &TransactionModel{
		ID:          t.ID,
		UserID:      t.UserID,
		Description: t.Description,
		Category:    t.Category,
		TotalAmount: t.TotalAmount,
		Date:        t.Date,
		CreatedAt:   t.CreatedAt,
		Items:       itemsData,
	}
}

/*
toTransactionItemModel maps a single domain.TransactionItem into its
GORM persistence representation.
*/
func toTransactionItemModel(item *domain.TransactionItem) TransactionItemModel {
	return TransactionItemModel{
		ID:            item.ID,
		TransactionID: item.TransactionID,
		Name:          item.Name,
		Quantity:      item.Quantity,
		Price:         item.Price,
	}
}

/*
toTransactionDomain maps a TransactionModel row, along with its
preloaded items, back into a domain.Transaction entity, keeping
GORM-specific types out of the business layer.
*/
func toTransactionDomain(m *TransactionModel) *domain.Transaction {
	var itemsData []domain.TransactionItem
	for _, item := range m.Items {
		itemsData = append(itemsData, toTransactionItemDomain(&item))
	}

	return &domain.Transaction{
		ID:          m.ID,
		UserID:      m.UserID,
		Description: m.Description,
		Category:    m.Category,
		TotalAmount: m.TotalAmount,
		Date:        m.Date,
		CreatedAt:   m.CreatedAt,
		Items:       itemsData,
	}
}

/*
toTransactionItemDomain maps a single TransactionItemModel row back
into a domain.TransactionItem entity.
*/
func toTransactionItemDomain(item *TransactionItemModel) domain.TransactionItem {
	return domain.TransactionItem{
		ID:            item.ID,
		TransactionID: item.TransactionID,
		Name:          item.Name,
		Quantity:      item.Quantity,
		Price:         item.Price,
	}
}

/*
Create persists a new transaction record along with its items. GORM
automatically inserts the related TransactionItemModel rows because
the relation is declared on TransactionModel.
*/
func (gtr *GormTransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	model := toTransactionModel(t)
	return gtr.db.WithContext(ctx).Create(model).Error
}

/*
FindByID looks up a transaction by its ID, preloading its items.
It returns (nil, nil) when no matching transaction exists,
distinguishing a "not found" result from an actual database error.
*/
func (gtr *GormTransactionRepository) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	var queryResult TransactionModel

	err := gtr.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&queryResult).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return toTransactionDomain(&queryResult), nil
}

/*
ListByUser returns all transactions belonging to the given user,
ordered by most recent first, with items preloaded for each.
*/
func (gtr *GormTransactionRepository) ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error) {
	var queryResults []TransactionModel

	err := gtr.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("date desc").
		Find(&queryResults).Error
	if err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(queryResults))
	for _, m := range queryResults {
		transactions = append(transactions, *toTransactionDomain(&m))
	}

	return transactions, nil
}

/*
Update persists changes to an existing transaction. Items are fully
replaced via GORM's association Replace mode: the old set of items is
cleared and swapped for the new one. Unscoped makes Replace delete the
removed items; without it GORM only nulls their TransactionID, which
leaves them orphaned in the database.
*/
func (gtr *GormTransactionRepository) Update(ctx context.Context, t *domain.Transaction) error {
	model := toTransactionModel(t)

	return gtr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(model).Association("Items").Unscoped().Replace(model.Items); err != nil {
			return err
		}
		return tx.Save(model).Error
	})
}

/*
Delete removes a transaction by ID. Related items are removed
automatically at the database level via the ON DELETE CASCADE
constraint declared on the Items relation.
*/
func (gtr *GormTransactionRepository) Delete(ctx context.Context, id string) error {
	return gtr.db.WithContext(ctx).Delete(&TransactionModel{}, "id = ?", id).Error
}
