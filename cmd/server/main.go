package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ssankrith/kart-backend/internal/api"
	"github.com/ssankrith/kart-backend/internal/catalog/memory"
	catalogpg "github.com/ssankrith/kart-backend/internal/catalog/supabase"
	"github.com/ssankrith/kart-backend/internal/config"
	"github.com/ssankrith/kart-backend/internal/domain"
	"github.com/ssankrith/kart-backend/internal/order"
	"github.com/ssankrith/kart-backend/internal/promo"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var cat domain.Catalog
	switch cfg.Catalog.Backend {
	case "memory":
		mem, err := memory.LoadFromFile(cfg.Catalog.Memory.ProductsPath)
		if err != nil {
			log.Fatalf("memory catalog: %v", err)
		}
		cat = mem
	case "supabase":
		dsn, err := cfg.DSN()
		if err != nil {
			log.Fatalf("supabase dsn: %v", err)
		}
		pool, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			log.Fatalf("pgx pool: %v", err)
		}
		defer pool.Close()
		cat = catalogpg.New(pool)
	default:
		log.Fatalf("unknown catalog backend %q", cfg.Catalog.Backend)
	}

	pc, err := promo.LoadFromGZIPFiles(promo.DirPaths(cfg.Promo.DataDir))
	if err != nil {
		log.Fatalf("promo corpora: %v", err)
	}

	svc := &order.Service{Catalog: cat, Promo: pc}
	h := &api.Handlers{Catalog: cat, Order: svc}
	router := api.NewRouter(h, cfg.Auth.APIKey)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		log.Printf("listening on %s (catalog=%s)", cfg.HTTP.Addr, cfg.Catalog.Backend)
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
