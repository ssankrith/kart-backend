package promo

import "testing"

func TestCouponCodePreludeOK(t *testing.T) {
	if !CouponCodePreludeOK("HAPPYHRS") {
		t.Fatal("expected 8-char ASCII ok")
	}
	if CouponCodePreludeOK("SHORT") {
		t.Fatal("expected too short false")
	}
	// 8 runes, multibyte first char → len(bytes) != rune count
	if CouponCodePreludeOK("€1234567") {
		t.Fatal("expected non-ASCII false")
	}
}
