package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// NewRouter registers routes on a gin Engine.
func NewRouter(h *Handlers, apiKey string) *gin.Engine {
	return NewRouterWithConfig(h, apiKey, RouterConfig{MaxBodyBytes: 65536})
}

// NewRouterWithConfig registers routes with optional rate limiting and body caps on order routes.
func NewRouterWithConfig(h *Handlers, apiKey string, cfg RouterConfig) *gin.Engine {
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(CORSMiddleware())

	register := func(r gin.IRoutes) {
		r.GET("/health", h.Health)
		r.GET("/ready", h.Readiness)
		r.GET("/product", h.ListProducts)
		r.GET("/product/:productId", h.GetProduct)
	}
	register(r)
	register(r.Group("/api"))

	var limiter *rate.Limiter
	if cfg.RateLimitRPS > 0 {
		burst := cfg.RateBurst
		if burst <= 0 {
			burst = int(cfg.RateLimitRPS * 2)
			if burst < 1 {
				burst = 1
			}
		}
		limiter = rate.NewLimiter(rate.Limit(cfg.RateLimitRPS), burst)
	}

	order := r.Group("")
	order.Use(OrderRateLimit(limiter))
	order.Use(MaxRequestBytes(cfg.MaxBodyBytes))
	order.Use(APIKeyAuth(apiKey))
	order.POST("/order", h.PlaceOrder)

	orderAPI := r.Group("/api")
	orderAPI.Use(OrderRateLimit(limiter))
	orderAPI.Use(MaxRequestBytes(cfg.MaxBodyBytes))
	orderAPI.Use(APIKeyAuth(apiKey))
	orderAPI.POST("/order", h.PlaceOrder)

	return r
}
