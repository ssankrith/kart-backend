package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ssankrith/kart-backend/internal/api"
	"github.com/ssankrith/kart-backend/internal/catalog/memory"
	"github.com/ssankrith/kart-backend/internal/observability"
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

func getenvFloat64(key string, def float64) float64 {
	loadEnvOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			log.Printf("godotenv: %v", err)
		}
	})
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("env %s: invalid float %q, using default %v", key, s, def)
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
	loadEnvOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			log.Printf("godotenv: %v", err)
		}
	})
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("env %s: invalid int %q, using default %d", key, s, def)
		return def
	}
	return v
}

func getenvInt64(key string, def int64) int64 {
	loadEnvOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			log.Printf("godotenv: %v", err)
		}
	})
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Printf("env %s: invalid int64 %q, using default %d", key, s, def)
		return def
	}
	return v
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
	rawPromo, err := promo.LoadPromo(couponDataDir)
	promoEnd := time.Now()
	if err != nil {
		log.Fatalf("promo corpus load: %v", err)
	}
	log.Printf("promo corpus loaded from %q: start %s, end %s, elapsed %s",
		couponDataDir,
		promoStart.Format(time.RFC3339Nano),
		promoEnd.Format(time.RFC3339Nano),
		promoEnd.Sub(promoStart))

	mc := observability.InstrumentPromo(rawPromo)
	defer func() { _ = mc.Close() }()

	svc := &order.Service{Catalog: cat, Promo: mc}
	h := &api.Handlers{Catalog: cat, Order: svc, Ready: func() bool { return true }}

	routerCfg := api.RouterConfig{
		RateLimitRPS: getenvFloat64("ORDER_RATE_RPS", 100),
		RateBurst:    getenvInt("ORDER_RATE_BURST", 200),
		MaxBodyBytes: getenvInt64("MAX_BODY_BYTES", 65536),
	}
	if routerCfg.RateLimitRPS <= 0 {
		log.Printf("ORDER_RATE_RPS<=0: rate limiting disabled")
	}
	router := api.NewRouterWithConfig(h, apiKey, routerCfg)

	metricsAddr := getenv("METRICS_ADDR", "")
	var metricsSrv *http.Server
	if metricsAddr != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		metricsSrv = &http.Server{Addr: metricsAddr, Handler: mux}
		go func() {
			log.Printf("metrics listening on %s", metricsAddr)
			if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("metrics: %v", err)
			}
		}()
	}

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
	if metricsSrv != nil {
		_ = metricsSrv.Shutdown(ctx)
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
