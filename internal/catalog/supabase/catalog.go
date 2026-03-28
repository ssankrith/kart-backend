package supabase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ssankrith/kart-backend/internal/catalog"
	"github.com/ssankrith/kart-backend/internal/domain"
)

// Catalog loads products from Postgres (Supabase-compatible).
type Catalog struct {
	pool *pgxpool.Pool
}

// New creates a catalog backed by pgx pool.
func New(pool *pgxpool.Pool) *Catalog {
	return &Catalog{pool: pool}
}

// List implements domain.Catalog.
func (c *Catalog) List(ctx context.Context) ([]domain.Product, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT id, name, price, category, image_json
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var out []domain.Product
	for rows.Next() {
		var p domain.Product
		var imgJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Category, &imgJSON); err != nil {
			return nil, err
		}
		if len(imgJSON) > 0 {
			var im domain.Image
			if err := json.Unmarshal(imgJSON, &im); err != nil {
				return nil, fmt.Errorf("image json: %w", err)
			}
			p.Image = &im
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Get implements domain.Catalog.
func (c *Catalog) Get(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	var imgJSON []byte
	err := c.pool.QueryRow(ctx, `
		SELECT id, name, price, category, image_json
		FROM products WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Price, &p.Category, &imgJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(imgJSON) > 0 {
		var im domain.Image
		if err := json.Unmarshal(imgJSON, &im); err != nil {
			return nil, fmt.Errorf("image json: %w", err)
		}
		p.Image = &im
	}
	return &p, nil
}
