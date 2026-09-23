package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
MaxBodySize returns a Gin middleware that caps the request body at
maxBytes. Reading past the limit makes the handler's bind call fail
with *http.MaxBytesError instead of buffering an arbitrarily large
body into memory.
*/
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
