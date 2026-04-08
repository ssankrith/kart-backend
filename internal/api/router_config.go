package api

// RouterConfig configures optional HTTP middleware (rate limits, body size).
type RouterConfig struct {
	// RateLimitRPS and RateBurst enable token-bucket limiting on order routes.
	// If RateLimitRPS <= 0, rate limiting is disabled.
	RateLimitRPS   float64
	RateBurst      int
	MaxBodyBytes   int64
}