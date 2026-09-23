package persistence

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

type GormUserRepository struct {
	db *gorm.DB
}

var _ repository.UserRepository = (*GormUserRepository)(nil)

/*
toUserModel maps a domain.User entity into its GORM persistence
representation (UserModel), so the database layer never depends
directly on the business entity's shape.
*/
func toUserModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	}
}

/*
toUserDomain maps a UserModel row fetched from the database back into
a domain.User entity, keeping GORM-specific types out of the business layer.
*/
func toUserDomain(u *UserModel) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	}
}

/*
NewGormUserRepository builds a GormUserRepository backed by the given
GORM connection. It satisfies the repository.UserRepository interface.
*/
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

/*
Create persists a new user record in the database.
It returns domain.ErrEmailAlreadyExists when the email's unique index
rejects the insert, which covers two concurrent registrations for the
same email slipping past the usecase's FindByEmail check.
*/
func (gru *GormUserRepository) Create(ctx context.Context, u *domain.User) error {
	model := toUserModel(u)

	err := gru.db.WithContext(ctx).Create(model).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrEmailAlreadyExists
	}

	return err
}

/*
FindByEmail looks up a user by email address.
It returns (nil, nil) when no matching user exists, distinguishing a
"not found" result from an actual database error.
*/
func (gru *GormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var queryResult UserModel

	err := gru.db.WithContext(ctx).Where("email = ?", email).First(&queryResult).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return toUserDomain(&queryResult), nil
}

/*
FindByID looks up a user by their unique identifier.
It returns (nil, nil) when no matching user exists, distinguishing a
"not found" result from an actual database error.
*/
func (gru *GormUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var queryResult UserModel

	err := gru.db.WithContext(ctx).Where("id = ?", id).First(&queryResult).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return toUserDomain(&queryResult), nil
}
