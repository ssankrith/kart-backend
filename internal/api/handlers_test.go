package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ssankrith/kart-backend/internal/catalog/memory"
	"github.com/ssankrith/kart-backend/internal/domain"
	"github.com/ssankrith/kart-backend/internal/order"
	"github.com/ssankrith/kart-backend/internal/promo"
)

func TestPlaceOrder_Unauthorized(t *testing.T) {
	cat := memory.NewFromSlice([]domain.Product{{ID: "1", Name: "A", Category: "c", Price: 1}})
	dir := couponDir(t)
	pc, err := promo.LoadFromGZIPFiles(promo.DirPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	h := &Handlers{Catalog: cat, Order: &order.Service{Catalog: cat, Promo: pc}}
	r := NewRouter(h, "secret")

	w := httptest.NewRecorder()
	body := `{"items":[{"productId":"1","quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
}

func TestPlaceOrder_OK(t *testing.T) {
	cat := memory.NewFromSlice([]domain.Product{{ID: "1", Name: "A", Category: "c", Price: 1}})
	dir := couponDir(t)
	pc, err := promo.LoadFromGZIPFiles(promo.DirPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	h := &Handlers{Catalog: cat, Order: &order.Service{Catalog: cat, Promo: pc}}
	r := NewRouter(h, "apitest")

	body := `{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHRS"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_key", "apitest")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var out OrderDTO
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.CouponCode != "HAPPYHRS" {
		t.Fatalf("coupon %+v", out)
	}
}

func couponDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeGZF(t, filepath.Join(dir, "couponbase1.gz"), "xx HAPPYHRS yy")
	writeGZF(t, filepath.Join(dir, "couponbase2.gz"), "HAPPYHRS")
	writeGZF(t, filepath.Join(dir, "couponbase3.gz"), "nope")
	return dir
}

func writeGZF(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}
