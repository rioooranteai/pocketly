package usecase

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"pocketly/internal/domain"
	"pocketly/internal/repository"
)

/*
AuthUsecase implements the business logic for user registration and
login. It depends only on interfaces — UserRepository, TokenSigner,
and PasswordHasher — never on their concrete implementations, so each
can be swapped independently without touching this file.
*/
type AuthUsecase struct {
	userRepo repository.UserRepository
	signer   repository.TokenSigner
	hasher   repository.PasswordHasher
}

/*
NewAuthUsecase builds an AuthUsecase backed by the given repository,
token signer, and password hasher.
*/
func NewAuthUsecase(userRepo repository.UserRepository, signer repository.TokenSigner, hasher repository.PasswordHasher) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, signer: signer, hasher: hasher}
}

/*
Register creates a new user account in the system.
It trims the name, normalizes the email, validates both, ensures email
uniqueness across the platform, and hashes the user's password via
PasswordHasher before persistence.
*/
func (uc *AuthUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	newUserData := &domain.User{
		Name:  strings.TrimSpace(name),
		Email: domain.NormalizeEmail(email),
	}

	if newUserData.Name == "" {
		return nil, domain.ErrInvalidName
	}

	if !newUserData.IsValidEmail() {
		return nil, domain.ErrInvalidEmail
	}

	existingUser, err := uc.userRepo.FindByEmail(ctx, newUserData.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	newUserData.ID = uuid.New().String()
	newUserData.CreatedAt = time.Now().UTC()

	hashedPassword, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	newUserData.Password = hashedPassword

	if err := uc.userRepo.Create(ctx, newUserData); err != nil {
		return nil, err
	}

	return newUserData, nil
}

/*
Login authenticates a user by email and password.
It returns a signed JWT along with the authenticated user's data on
success. To prevent user enumeration, a non-existent email, a wrong
password, and a hasher failure all result in the same
domain.ErrInvalidCredentials error. A non-existent email still runs
one password hash, so its response time matches a wrong password and
cannot be used to tell which emails are registered.
*/
func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	userData, err := uc.userRepo.FindByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		return "", nil, err
	}
	if userData == nil {
		_, _ = uc.hasher.Hash(password)
		return "", nil, domain.ErrInvalidCredentials
	}

	valid, err := uc.hasher.Verify(password, userData.Password)
	if err != nil {
		log.Printf("password verify failed for user %s: %v", userData.ID, err)
		return "", nil, domain.ErrInvalidCredentials
	}
	if !valid {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := uc.signer.Sign(userData.ID)
	if err != nil {
		return "", nil, err
	}

	return token, userData, nil
}
