package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pocketly/internal/infrastructure/auth"
)

/*
AuthMiddleware returns a Gin middleware that enforces JWT
authentication. It expects an "Authorization: Bearer <token>" header,
validates the token using the given signer, and stores the extracted
user ID in the request context under the key "userID" for downstream
handlers to use. Requests with a missing or invalid token are aborted
with 401 Unauthorized before reaching any handler.
*/
func AuthMiddleware(signer *auth.JWTSigner) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := signer.ParseUserID(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", userID)

		c.Next()
	}
}
