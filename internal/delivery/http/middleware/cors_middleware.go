package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
CORS returns a Gin middleware that lets browsers on any origin call
the API. This is meant for local development only — before deploying,
restrict Access-Control-Allow-Origin to the real frontend origin(s).
"*" works here because auth uses a Bearer token rather than cookies.

Browsers send an OPTIONS preflight before cross-origin requests that
carry a JSON body or an Authorization header. No route handles OPTIONS,
so this middleware answers it with 204 itself — it must be registered
with router.Use so it also runs for requests that match no route.
*/
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Expose-Headers", "Retry-After")
		c.Header("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
