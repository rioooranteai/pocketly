package main

import (
	"log"

	"github.com/gin-gonic/gin"

	router "pocketly/internal/delivery/http"
	"pocketly/internal/delivery/http/handler"
	"pocketly/internal/infrastructure/auth"
	"pocketly/internal/infrastructure/config"
	persistence "pocketly/internal/infrastructure/persistence/gorm"
	"pocketly/internal/usecase"
)

/*
main is the application entry point. It wires every layer together
in dependency order — configuration, database connection, repository,
token signer, usecase, HTTP handler, and router — then starts the
HTTP server. This is the only place in the codebase allowed to know
about every layer at once; every other package depends only on the
interfaces or types it directly needs.
*/
func main() {
	cfg := config.Load()

	db, err := persistence.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	userRepository := persistence.NewGormUserRepository(db)

	jwtSigner := auth.NewJWTSigner(cfg.JWTSecret)

	authUsecase := usecase.NewAuthUseCase(userRepository, jwtSigner)

	authHandler := handler.NewAuthHandler(authUsecase)

	r := gin.Default()

	router.SetupRoutes(r, authHandler)

	r.Run(": " + cfg.AppPort)
}
