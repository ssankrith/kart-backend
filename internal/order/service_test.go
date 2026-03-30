package order

import (
	"context"
	"errors"
	"testing"

	"github.com/ssankrith/kart-backend/internal/catalog/memory"
	"github.com/ssankrith/kart-backend/internal/domain"
)

type fakePromo struct {
	ok bool
}

func (f fakePromo) Valid(string) bool { return f.ok }

func (fakePromo) Close() error { return nil }

func TestPlace_InvalidProduct(t *testing.T) {
	cat := memory.NewFromSlice([]domain.Product{{ID: "1", Name: "A", Price: 1, Category: "c"}})
	s := &Service{Catalog: cat, Promo: fakePromo{ok: true}}
	_, err := s.Place(context.Background(), []Line{{ProductID: "99", Quantity: 1}}, nil)
	if !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("got %v", err)
	}
}

func TestPlace_InvalidCoupon(t *testing.T) {
	cat := memory.NewFromSlice([]domain.Product{{ID: "1", Name: "A", Price: 1, Category: "c"}})
	s := &Service{Catalog: cat, Promo: fakePromo{ok: false}}
	code := "HAPPYHRS"
	_, err := s.Place(context.Background(), []Line{{ProductID: "1", Quantity: 1}}, &code)
	if !errors.Is(err, ErrInvalidCoupon) {
		t.Fatalf("got %v", err)
	}
}

func TestPlace_OK(t *testing.T) {
	cat := memory.NewFromSlice([]domain.Product{{ID: "1", Name: "A", Price: 1, Category: "c"}})
	s := &Service{Catalog: cat, Promo: fakePromo{ok: true}}
	code := "HAPPYHRS"
	res, err := s.Place(context.Background(), []Line{{ProductID: "1", Quantity: 2}}, &code)
	if err != nil {
		t.Fatal(err)
	}
	if res.CouponCode != code {
		t.Fatalf("coupon %q", res.CouponCode)
	}
}
