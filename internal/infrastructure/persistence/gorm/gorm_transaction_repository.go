package persistence

import (
	"context"

	"gorm.io/gorm"

	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

type GormTransactionRepository struct {
	db *gorm.DB
}

var _ repository.TransactionRepository = (*GormTransactionRepository)(nil)

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

func toTransactionItemModel(item *domain.TransactionItem) TransactionItemModel {
	return TransactionItemModel{
		ID:            item.ID,
		TransactionID: item.TransactionID,
		Name:          item.Name,
		Quantity:      item.Quantity,
		Price:         item.Price,
	}
}

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

func toTransactionItemDomain(item *TransactionItemModel) domain.TransactionItem {
	return domain.TransactionItem{
		ID:            item.ID,
		TransactionID: item.TransactionID,
		Name:          item.Name,
		Quantity:      item.Quantity,
		Price:         item.Price,
	}
}

func (gtr *GormTransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	panic("not implemented")
}

func (gtr *GormTransactionRepository) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	panic("not implemented")
}

func (gtr *GormTransactionRepository) ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error) {
	panic("not implemented")
}

func (gtr *GormTransactionRepository) Update(ctx context.Context, t *domain.Transaction) error {
	panic("not implemented")
}

func (gtr *GormTransactionRepository) Delete(ctx context.Context, id string) error {
	panic("not implemented")
}
