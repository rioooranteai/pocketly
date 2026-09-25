package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"pocketly/internal/delivery/http/handler"
	"pocketly/internal/delivery/http/middleware"
	"pocketly/internal/port"
)

/*
authBodyLimit caps register and login request bodies. Their largest
valid payload is well under 1KB, so 4KB leaves room without letting
a client stream megabytes into the JSON decoder.
*/
const authBodyLimit = 4 * 1024

/*
transactionBodyLimit caps create and update transaction bodies. A
receipt with a hundred line items is around 10KB of JSON, so 64KB is
generous while still stopping a client from streaming a huge body
into the JSON decoder.
*/
const transactionBodyLimit = 64 * 1024

/*
scanMultipartOverhead is added on top of the image size limit for
/transactions/scan, leaving room for the multipart boundaries and part
headers around the file so an image right at the limit still fits.
*/
const scanMultipartOverhead = 64 * 1024

/*
SetupRoutes registers all HTTP routes for the application and wires
each one to its corresponding handler method. This is the single
place where URL paths and HTTP methods are mapped to application
behavior — the handlers themselves stay unaware of routing details.
Transaction routes are protected by AuthMiddleware; auth routes
(register, login) remain public since they are used before a user
has a token at all, so they are rate limited and size limited instead.
maxImageSize caps the receipt upload on /transactions/scan, so an
oversized body is rejected while it is read instead of being buffered
into memory first.
*/
func SetupRoutes(router *gin.Engine, authHandler *handler.AuthHandler, transactionHandler *handler.TransactionHandler, signer port.TokenVerifier, maxImageSize int64) {
	v1 := router.Group("/api/v1")

	{
		/*
			Public routes: no token required, since these are the entry
			points a user goes through before they have one. Each route
			has its own per-IP limit to slow down password guessing and
			mass account creation, and every attempt costs an Argon2 hash.
		*/
		v1.POST("/register", middleware.RateLimit(5, time.Minute), middleware.MaxBodySize(authBodyLimit), authHandler.Register)
		v1.POST("/login", middleware.RateLimit(10, time.Minute), middleware.MaxBodySize(authBodyLimit), authHandler.Login)

		/*
			Protected routes: AuthMiddleware runs first on every request
			here. It aborts with 401 before the handler runs if the token
			is missing or invalid; otherwise it stores the authenticated
			user's ID in the request context for the handler to read.
		*/
		v1.POST("/transactions", middleware.AuthMiddleware(signer), middleware.MaxBodySize(transactionBodyLimit), transactionHandler.Create)
		v1.GET("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Get)
		v1.GET("/transactions", middleware.AuthMiddleware(signer), transactionHandler.List)
		v1.PUT("/transactions/:id", middleware.AuthMiddleware(signer), middleware.MaxBodySize(transactionBodyLimit), transactionHandler.Update)
		v1.DELETE("/transactions/:id", middleware.AuthMiddleware(signer), transactionHandler.Delete)
		v1.POST("/transactions/scan", middleware.AuthMiddleware(signer), middleware.MaxBodySize(maxImageSize+scanMultipartOverhead), transactionHandler.Scan)
	}

}
