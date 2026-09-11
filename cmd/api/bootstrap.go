package main

import (
	"gorm.io/gorm"

	"pocketly/internal/delivery/http/handler"
	"pocketly/internal/infrastructure/ai"
	"pocketly/internal/infrastructure/auth"
	"pocketly/internal/infrastructure/config"
	persistence "pocketly/internal/infrastructure/persistence/gorm"
	"pocketly/internal/usecase"
)

/*
App holds every fully-wired handler the router needs. It is the
single object main.go depends on, so main.go itself stays tiny and
free of construction details.
*/
type App struct {
	AuthHandler        *handler.AuthHandler
	TransactionHandler *handler.TransactionHandler
	JWTSigner          *auth.JWTSigner
}

/*
Bootstrap wires every layer together in dependency order —
repositories, signers, categorizer, usecases, then handlers — and
returns a ready-to-use App. This is the only function in the
codebase allowed to know about every concrete implementation at
once; everything it constructs is handed to consumers only through
the interfaces they depend on.
*/
func Bootstrap(cfg config.Config, db *gorm.DB) (*App, error) {
	userRepository := persistence.NewGormUserRepository(db)
	jwtSigner := auth.NewJWTSigner(cfg.JWTSecret)
	authUsecase := usecase.NewAuthUsecase(userRepository, jwtSigner)
	authHandler := handler.NewAuthHandler(authUsecase)

	categorizer, err := ai.NewCategorizer(cfg)
	if err != nil {
		return nil, err
	}

	transactionRepository := persistence.NewGormTransactionRepository(db)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepository, categorizer)
	transactionHandler := handler.NewTransactionHandler(transactionUsecase)

	return &App{
		AuthHandler:        authHandler,
		TransactionHandler: transactionHandler,
		JWTSigner:          jwtSigner,
	}, nil
}
