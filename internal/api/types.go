package api

import "github.com/ssankrith/kart-backend/internal/domain"

// OrderLineIn is the request line item (OpenAPI OrderReq).
type OrderLineIn struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

// OrderReq is POST /order body.
type OrderReq struct {
	Items      []OrderLineIn `json:"items" binding:"required,dive"`
	CouponCode *string       `json:"couponCode,omitempty"`
}

// OrderItemOut is each element of Order.items in responses.
type OrderItemOut struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// ProductDTO is the wire format for Product.
type ProductDTO struct {
	ID       string    `json:"id"`
	Image    *ImageDTO `json:"image,omitempty"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
}

// ImageDTO matches demo nested image object.
type ImageDTO struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}

// OrderDTO is POST /order success body.
type OrderDTO struct {
	ID         string         `json:"id"`
	Items      []OrderItemOut `json:"items"`
	CouponCode string         `json:"couponCode,omitempty"`
	Products   []ProductDTO   `json:"products"`
}

// ErrorBody is used for 4xx JSON errors.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ProductToDTO maps domain → wire.
func ProductToDTO(p domain.Product) ProductDTO {
	out := ProductDTO{
		ID:       p.ID,
		Name:     p.Name,
		Category: p.Category,
		Price:    p.Price,
	}
	if p.Image != nil {
		out.Image = &ImageDTO{
			Thumbnail: p.Image.Thumbnail,
			Mobile:    p.Image.Mobile,
			Tablet:    p.Image.Tablet,
			Desktop:   p.Image.Desktop,
		}
	}
	return out
}
