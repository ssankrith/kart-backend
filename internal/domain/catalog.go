package domain

import "context"

// Catalog lists and resolves products for pricing and order assembly.
type Catalog interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id string) (*Product, error)
}
