package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
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

// OrderRateLimit rejects requests when the shared limiter has no tokens.
func OrderRateLimit(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter != nil && !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorBody{
				Code:    "rate_limited",
				Message: "too many requests",
			})
			return
		}
		c.Next()
	}
}

// MaxRequestBytes limits JSON body size (e.g. for POST /order).
func MaxRequestBytes(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if max <= 0 {
			c.Next()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}
