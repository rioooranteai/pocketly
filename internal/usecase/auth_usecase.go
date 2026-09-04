package usecase

import (
	"context"
	"time"

	"pocketly/internal/domain"
	"pocketly/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo repository.UserRepository
	signer   repository.TokenSigner
}

func NewAuthUsecase(userRepo repository.UserRepository, signer repository.TokenSigner) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, signer: signer}
}

/*
Register creates a new user account in the system.
It validates the email format, ensures email uniqueness across the platform,
and applies bcrypt hashing to the user's password before persistence.
*/
func (uc *AuthUsecase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	newUserData := &domain.User{
		Name:  name,
		Email: email,
	}

	if !newUserData.IsValidEmail() {
		return nil, domain.ErrInvalidEmail
	}

	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	newUserData.ID = uuid.New().String()
	newUserData.CreatedAt = time.Now().UTC()

	// Cost is set to DefaultCost (10). If performance becomes a bottleneck
	// during traffic spikes, consider offloading this to a background worker.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUserData.Password = string(hashedPassword)

	err = uc.userRepo.Create(ctx, newUserData)
	if err != nil {
		return nil, err
	}

	return newUserData, nil
}

/*
Login authenticates a user by email and password.

	It returns a signed JWT along with the authenticated user's data on success.
	To prevent user enumeration, both a non-existent email and a wrong password
	result in the same domain.ErrInvalidCredentials error.
*/
func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	userData, err := uc.userRepo.FindByEmail(ctx, email)

	if err != nil {
		return "", nil, err
	}
	if userData == nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(password))

	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := uc.signer.Sign(userData.ID)

	if err != nil {
		return "", nil, err
	}

	return token, userData, nil
}
