package repository

import (
	"context"
)

type CategorizerRepository interface {
	Categorize(ctx context.Context, description string) (string, error)
}
