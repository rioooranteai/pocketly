package router

import (
	"github.com/gin-gonic/gin"

	"pocketly/internal/delivery/http/handler"
)

/*
SetupRoutes registers all HTTP routes for the application and wires
each one to its corresponding handler method. This is the single
place where URL paths and HTTP methods are mapped to application
behavior — the handlers themselves stay unaware of routing details.
*/
func SetupRoutes(router *gin.Engine, authHandler *handler.AuthHandler) {
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
}
