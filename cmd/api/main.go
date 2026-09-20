package main

import (
	"log"

	"github.com/gin-gonic/gin"

	router "pocketly/internal/delivery/http"
	"pocketly/internal/infrastructure/config"
	persistence "pocketly/internal/infrastructure/persistence/gorm"
)

/*
main loads configuration, opens the database connection, delegates
all dependency wiring to Bootstrap, and starts the HTTP server. It
intentionally contains no construction logic itself — see
bootstrap.go for that.
*/
func main() {
	cfg := config.Load()

	db, err := persistence.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app, err := Bootstrap(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	router.SetupRoutes(r, app.AuthHandler, app.TransactionHandler, app.JWTSigner)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
