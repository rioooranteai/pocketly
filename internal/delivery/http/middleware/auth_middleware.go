package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pocketly/internal/port"
)

/*
UserIDKey is the gin.Context key under which AuthMiddleware stores the
authenticated user's ID. Handlers and other middleware read it through
this constant rather than a repeated string literal, so a typo cannot
silently turn every request into 401.
*/
const UserIDKey = "userID"

/*
AuthMiddleware returns a Gin middleware that enforces JWT
authentication. It expects an "Authorization: Bearer <token>" header
(the scheme is matched case-insensitively, as HTTP requires),
validates the token using the given verifier, and stores the extracted
user ID in the request context under UserIDKey for downstream
handlers to use. Requests with a missing, malformed, or invalid token
are aborted with 401 Unauthorized before reaching any handler.
*/
func AuthMiddleware(verifier port.TokenVerifier) gin.HandlerFunc {
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

		c.Set(UserIDKey, userID)

		c.Next()
	}
}
