package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ssankrith/kart-backend/internal/catalog"
	"github.com/ssankrith/kart-backend/internal/domain"
	"github.com/ssankrith/kart-backend/internal/order"
)

// Handlers holds HTTP dependencies.
type Handlers struct {
	Catalog domain.Catalog
	Order   *order.Service
	// Ready returns true when the process can accept traffic (catalog + promo loaded).
	// If nil, readiness probes always succeed.
	Ready func() bool
}

// Health is liveness: process is running.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness returns 200 when dependencies are satisfied, else 503.
func (h *Handlers) Readiness(c *gin.Context) {
	if h.Ready != nil && !h.Ready() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// ListProducts GET /product
func (h *Handlers) ListProducts(c *gin.Context) {
	items, err := h.Catalog.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorBody{Code: "internal", Message: err.Error()})
		return
	}
	out := make([]ProductDTO, 0, len(items))
	for _, p := range items {
		out = append(out, ProductToDTO(p))
	}
	c.JSON(http.StatusOK, out)
}

// GetProduct GET /product/:productId
func (h *Handlers) GetProduct(c *gin.Context) {
	id := c.Param("productId")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorBody{Code: "bad_request", Message: "missing product id"})
		return
	}
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, ErrorBody{Code: "bad_request", Message: "invalid product id"})
		return
	}
	p, err := h.Catalog.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorBody{Code: "internal", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ProductToDTO(*p))
}

// PlaceOrder POST /order
func (h *Handlers) PlaceOrder(c *gin.Context) {
	var req OrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorBody{Code: "bad_request", Message: err.Error()})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusUnprocessableEntity, ErrorBody{Code: "validation", Message: "at least one item is required"})
		return
	}
	lines := make([]order.Line, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, order.Line{ProductID: it.ProductID, Quantity: it.Quantity})
	}
	res, err := h.Order.Place(c.Request.Context(), lines, req.CouponCode)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrEmptyItems):
			c.JSON(http.StatusUnprocessableEntity, ErrorBody{Code: "validation", Message: "at least one item is required"})
		case errors.Is(err, order.ErrInvalidQty):
			c.JSON(http.StatusUnprocessableEntity, ErrorBody{Code: "validation", Message: "invalid quantity"})
		case errors.Is(err, order.ErrInvalidProduct):
			c.JSON(http.StatusUnprocessableEntity, ErrorBody{Code: "constraint", Message: "invalid product specified"})
		case errors.Is(err, order.ErrInvalidCoupon):
			c.JSON(http.StatusUnprocessableEntity, ErrorBody{Code: "validation", Message: "invalid coupon code"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorBody{Code: "internal", Message: err.Error()})
		}
		return
	}
	itemsOut := make([]OrderItemOut, 0, len(res.Lines))
	for _, ln := range res.Lines {
		itemsOut = append(itemsOut, OrderItemOut{ProductID: ln.ProductID, Quantity: ln.Quantity})
	}
	prods := make([]ProductDTO, 0, len(res.Products))
	for _, p := range res.Products {
		prods = append(prods, ProductToDTO(p))
	}
	c.JSON(http.StatusOK, OrderDTO{
		ID:         res.ID.String(),
		Items:      itemsOut,
		CouponCode: res.CouponCode,
		Products:   prods,
	})
}
