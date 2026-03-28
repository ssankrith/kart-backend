package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth requires header api_key to match expected (POST /order).
func APIKeyAuth(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("api_key"))
		if key == "" {
			key = strings.TrimSpace(c.GetHeader("X-API-Key"))
		}
		if key != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorBody{
				Code:    "unauthorized",
				Message: "missing or invalid api_key",
			})
			return
		}
		c.Next()
	}
}
