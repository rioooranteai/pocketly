package repository

import (
	"context"

	"pocketly/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

