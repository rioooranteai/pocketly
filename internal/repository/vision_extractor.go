package repository

import (
	"context"
	"pocketly/internal/domain"
)

type VisionExtractor interface {
	Extract(ctx context.Context, imageData []byte) (description string, items []domain.TransactionItem, err error)
}
