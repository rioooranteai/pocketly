package repository

import (
	"context"

	"pocketly/internal/domain"
)

/*
UserRepository is the contract for persisting and retrieving user
accounts. The usecase layer depends only on this interface, never on
a concrete implementation, so the underlying storage (SQLite, Postgres,
or anything else) can be swapped without touching business logic.
*/
type UserRepository interface {
	/*
		Create persists a new user record.
	*/
	Create(ctx context.Context, u *domain.User) error

	/*
		FindByEmail looks up a user by email address.
		Implementations should return (nil, nil) when no user matches,
		distinguishing "not found" from an actual error.
	*/
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	/*
		FindByID looks up a user by their unique identifier.
		Implementations should return (nil, nil) when no user matches,
		distinguishing "not found" from an actual error.
	*/
	FindByID(ctx context.Context, id string) (*domain.User, error)
}
