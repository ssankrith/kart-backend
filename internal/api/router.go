package api

import (
	"os"

	"github.com/gin-gonic/gin"
)

// NewRouter registers routes on a gin Engine.
func NewRouter(h *Handlers, apiKey string) *gin.Engine {
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	register := func(r gin.IRoutes) {
		r.GET("/health", h.Health)
		r.GET("/product", h.ListProducts)
		r.GET("/product/:productId", h.GetProduct)
	}
	register(r)
	register(r.Group("/api"))

	order := r.Group("")
	order.Use(APIKeyAuth(apiKey))
	order.POST("/order", h.PlaceOrder)

	orderAPI := r.Group("/api")
	orderAPI.Use(APIKeyAuth(apiKey))
	orderAPI.POST("/order", h.PlaceOrder)

	return r
}
