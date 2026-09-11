package router

import (
	"github.com/gin-gonic/gin"

	"pocketly/internal/delivery/http/handler"
	"pocketly/internal/delivery/http/middleware"
	"pocketly/internal/infrastructure/auth"
)

/*
SetupRoutes registers all HTTP routes for the application and wires
each one to its corresponding handler method. This is the single
place where URL paths and HTTP methods are mapped to application
behavior — the handlers themselves stay unaware of routing details.
Transaction routes are protected by AuthMiddleware; auth routes
(register, login) remain public since they are used before a user
has a token at all.
*/
func SetupRoutes(router *gin.Engine, authHandler *handler.AuthHandler, transactionHandler *handler.TransactionHandler, signer *auth.JWTSigner) {
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	router.POST("/transactions", middleware.AuthMiddleware(signer), transactionHandler.Create)
	router.GET("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Get)
	router.GET("/transactions", middleware.AuthMiddleware(signer), transactionHandler.List)
	router.PUT("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Update)
	router.DELETE("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Delete)
}
