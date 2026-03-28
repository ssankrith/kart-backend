package domain

// Product matches the demo API shape (OpenAPI Product + optional image URLs).
type Product struct {
	ID       string `json:"id"`
	Image    *Image `json:"image,omitempty"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    float64 `json:"price"`
}

// Image holds responsive asset URLs from the demo storefront.
type Image struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}
