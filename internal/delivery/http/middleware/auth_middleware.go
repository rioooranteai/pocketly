package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pocketly/internal/repository"
)

/*
AuthMiddleware returns a Gin middleware that enforces JWT
authentication. It expects an "Authorization: Bearer <token>" header
(the scheme is matched case-insensitively, as HTTP requires),
validates the token using the given verifier, and stores the extracted
user ID in the request context under the key "userID" for downstream
handlers to use. Requests with a missing, malformed, or invalid token
are aborted with 401 Unauthorized before reaching any handler.
*/
func AuthMiddleware(verifier repository.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		scheme, token, found := strings.Cut(authHeader, " ")
		token = strings.TrimSpace(token)
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be in the format: Bearer <token>"})
			return
		}

		userID, err := verifier.ParseUserID(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", userID)

		c.Next()
	}
}
