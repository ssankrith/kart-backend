package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ssankrith/kart-backend/internal/catalog"
	"github.com/ssankrith/kart-backend/internal/domain"
)

// Catalog is an in-memory product store loaded from JSON.
type Catalog struct {
	byID map[string]domain.Product
	list []domain.Product
}

// LoadFromFile reads JSON array of products (demo format).
func LoadFromFile(path string) (*Catalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []domain.Product
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, fmt.Errorf("parse products json: %w", err)
	}
	return NewFromSlice(items), nil
}

// NewFromSlice builds a catalog from an in-memory slice (tests).
func NewFromSlice(items []domain.Product) *Catalog {
	byID := make(map[string]domain.Product, len(items))
	for _, p := range items {
		byID[p.ID] = p
	}
	return &Catalog{byID: byID, list: append([]domain.Product(nil), items...)}
}

// List implements domain.Catalog.
func (c *Catalog) List(_ context.Context) ([]domain.Product, error) {
	out := make([]domain.Product, len(c.list))
	copy(out, c.list)
	return out, nil
}

// Get implements domain.Catalog.
func (c *Catalog) Get(_ context.Context, id string) (*domain.Product, error) {
	p, ok := c.byID[id]
	if !ok {
		return nil, catalog.ErrNotFound
	}
	return &p, nil
}
