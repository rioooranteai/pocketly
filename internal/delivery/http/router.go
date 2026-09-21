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
	/*
		Public routes: no token required, since these are the entry
		points a user goes through before they have one.
	*/
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	/*
		Protected routes: AuthMiddleware runs first on every request
		here. It aborts with 401 before the handler runs if the token
		is missing or invalid; otherwise it stores the authenticated
		user's ID in the request context for the handler to read.
	*/
	router.POST("/transactions", middleware.AuthMiddleware(signer), transactionHandler.Create)
	router.GET("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Get)
	router.GET("/transactions", middleware.AuthMiddleware(signer), transactionHandler.List)
	router.PUT("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Update)
	router.DELETE("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Delete)
	router.POST("/transactions/scan", middleware.AuthMiddleware(signer), transactionHandler.Scan)
}
