# kart-backend

Go implementation of the [Oolio advanced backend challenge](https://github.com/oolio-group/kart-challenge/blob/advanced-challenge/backend-challenge/README.md): food-ordering API per [OpenAPI 3.1](https://orderfoodonline.deno.dev/public/openapi.yaml), with **Gin**, configurable **in-memory or Supabase/Postgres** catalog, and **promo validation** against three gzipped corpora.

## Features

- `GET /product` or `GET /api/product` — list products (same as [demo](https://orderfoodonline.deno.dev/api) `/api` prefix)  
- `GET /product/{productId}` — product by id (400 invalid id, 404 missing)  
- `POST /order` or `POST /api/order` — place order (requires `api_key` header; validates optional `couponCode` against Oolio rules)  
- `GET /health` — liveness  

**Promo rules:** UTF-8 length 8–10 inclusive; code must appear as a substring in **at least two** of `couponbase1.gz` … `couponbase3.gz` under `promo.data_dir` (or `COUPON_DATA_DIR`).

## Quick start

```bash
cp config.example.yaml config.yaml
# data/products.json and data/couponbase*.gz are committed for local dev (small coupon fixtures).
# For production coupons, run: make fetch-coupons
export CONFIG_PATH=config.yaml
go run ./cmd/server
```

```bash
curl -s localhost:8080/product | head -c 200
curl -s -H 'api_key: apitest' -H 'Content-Type: application/json' \
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHRS"}' \
  localhost:8080/order
```

## Configuration

| Env / YAML | Purpose |
|------------|---------|
| `CONFIG_PATH` | Path to YAML (default `config.yaml`) |
| `HTTP_ADDR` | Listen address (default `:8080`) |
| `API_KEY` | Required value for header `api_key` on `POST /order` |
| `catalog.backend` | `memory` or `supabase` |
| `catalog.memory.products_path` | JSON array of products (demo shape) |
| `DATABASE_URL` | Postgres URL when `catalog.backend: supabase` |
| `COUPON_DATA_DIR` | Directory with `couponbase1.gz` … `3` |

### Supabase

1. Run `migrations/001_products.up.sql` in the Supabase SQL editor (or `psql`).  
2. Load rows from `data/products.json` (map `image` object to `image_json` JSONB).  
3. Set `catalog.backend: supabase` and `DATABASE_URL`.

## Docker

```bash
make docker
docker run --rm -p 8080:8080 kart-backend:local
```

## CI

GitHub Actions runs `go test ./... -race` and `go build` on push/PR (see `.github/workflows/ci.yml`).

## Conventional commits

Examples: `feat(api): …`, `fix(promo): …`, `test(order): …`, `docs: …`, `chore: …`.

## Scaling (notes)

- Run **multiple stateless** replicas behind a load balancer; keep sessions out of process memory.  
- Use **Supabase pooler** URL and bounded **pgx** pool settings for Postgres.  
- Large coupon files: corpora are loaded **once** at startup; at very large scale move indexing to a dedicated service or object storage.

## License

MIT (or your choice) — challenge submission.
