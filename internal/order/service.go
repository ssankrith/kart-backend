package order

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/ssankrith/kart-backend/internal/catalog"
	"github.com/ssankrith/kart-backend/internal/domain"
)

// Service builds validated orders from catalog + promo rules.
type Service struct {
	Catalog domain.Catalog
	Promo   domain.PromoChecker
}

// Line is one request line after validation.
type Line struct {
	ProductID string
	Quantity  int
}

// Result is the data needed to render OrderDTO.
type Result struct {
	ID         uuid.UUID
	Lines      []Line
	CouponCode string // set only when client sent a code and it was valid
	Products   []domain.Product
}

var (
	ErrEmptyItems     = errors.New("at least one item is required")
	ErrInvalidProduct = errors.New("invalid product specified")
	ErrInvalidQty     = errors.New("invalid quantity")
	ErrInvalidCoupon  = errors.New("invalid coupon code")
)

// Place validates input and returns order data or a domain error.
func (s *Service) Place(ctx context.Context, lines []Line, coupon *string) (*Result, error) {
	if len(lines) == 0 {
		return nil, ErrEmptyItems
	}
	for _, ln := range lines {
		if ln.Quantity <= 0 {
			return nil, fmt.Errorf("%w", ErrInvalidQty)
		}
		if ln.ProductID == "" {
			return nil, fmt.Errorf("%w", ErrInvalidProduct)
		}
	}

	if coupon != nil && *coupon != "" {
		if !s.Promo.Valid(*coupon) {
			return nil, fmt.Errorf("%w", ErrInvalidCoupon)
		}
	}

	// Unique product ids in stable order (sorted) to match demo "products" array shape.
	seen := make(map[string]struct{})
	var ids []string
	for _, ln := range lines {
		if _, ok := seen[ln.ProductID]; ok {
			continue
		}
		seen[ln.ProductID] = struct{}{}
		ids = append(ids, ln.ProductID)
	}
	sort.Strings(ids)

	products := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		p, err := s.Catalog.Get(ctx, id)
		if err != nil {
			if errors.Is(err, catalog.ErrNotFound) {
				return nil, fmt.Errorf("%w", ErrInvalidProduct)
			}
			return nil, err
		}
		products = append(products, *p)
	}

	out := &Result{
		ID:       uuid.New(),
		Lines:    lines,
		Products: products,
	}
	if coupon != nil && *coupon != "" {
		out.CouponCode = *coupon
	}
	return out, nil
}
