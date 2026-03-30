package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ssankrith/kart-backend/internal/api"
	"github.com/ssankrith/kart-backend/internal/catalog/memory"
	"github.com/ssankrith/kart-backend/internal/order"
	"github.com/ssankrith/kart-backend/internal/promo"
)

// loadEnvOnce loads a `.env` file from the working directory into the process
// environment (if present). Existing OS env vars are not overridden.
var loadEnvOnce sync.Once

func getenv(key, def string) string {
	loadEnvOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			log.Printf("godotenv: %v", err)
		}
	})
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	addr := getenv("HTTP_ADDR", ":8080")
	apiKey := getenv("API_KEY", "apitest")
	productsPath := getenv("PRODUCTS_PATH", "data/products.json")
	couponDataDir := getenv("COUPON_DATA_DIR", "data")

	cat, err := memory.LoadFromFile(productsPath)
	if err != nil {
		log.Fatalf("products: %v", err)
	}

	promoStart := time.Now()
	mc, err := promo.LoadPromo(couponDataDir)
	promoEnd := time.Now()
	if err != nil {
		log.Fatalf("promo corpus load: %v", err)
	}
	log.Printf("promo corpus loaded from %q: start %s, end %s, elapsed %s",
		couponDataDir,
		promoStart.Format(time.RFC3339Nano),
		promoEnd.Format(time.RFC3339Nano),
		promoEnd.Sub(promoStart))
	defer mc.Close()
	svc := &order.Service{Catalog: cat, Promo: mc}
	h := &api.Handlers{Catalog: cat, Order: svc}
	router := api.NewRouter(h, apiKey)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
