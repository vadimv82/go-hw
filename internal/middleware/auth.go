package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware creates a middleware that validates X-API-Key header
// Returns 401 Unauthorized if header is missing
// Returns 403 Forbidden if header value is incorrect
func APIKeyAuthMiddleware() gin.HandlerFunc {
	expectedAPIKey := os.Getenv("API_KEY")
	if expectedAPIKey == "" {
		// If no API key is configured, allow all requests (for development)
		// In production, you should set API_KEY environment variable
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-API-Key header",
			})
			c.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
