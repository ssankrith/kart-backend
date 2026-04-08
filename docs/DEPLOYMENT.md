# Deployment guide

## Prerequisites

- Go toolchain (see `go.mod` for version).
- Three gzip corpora: `data/couponbase1.gz`, `data/couponbase2.gz`, `data/couponbase3.gz`.
- Product catalog JSON: default `data/products.json`.
- Built shard directory: `000.bin` … `255.bin` produced by `cmd/preprocessor_seq`.

## Environment variables

| Variable | Typical value | Notes |
|----------|----------------|-------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `API_KEY` | strong secret in prod | Required header `api_key` on `POST /order` |
| `PRODUCTS_PATH` | `./data/products.json` | Path to product JSON array |
| `COUPON_DATA_DIR` | `./data` | Directory containing `couponbase*.gz` (used for paths/logging; shards are separate) |
| `PROMO_SHARDS_DIR` | `./shards_seq` | Directory with `000.bin` … `255.bin` |
| `PROMO_SHARDS_STRICT` | unset | Set `1` to require `manifest.json` and fail if shard files do not match recorded sizes/checksums. |
| `METRICS_ADDR` | unset | e.g. `:9091` — Prometheus `/metrics` on a **separate** listener (recommended for public hosts). |
| `ORDER_RATE_RPS` / `ORDER_RATE_BURST` | `100` / `200` | Rate limit on `POST /order`; use `ORDER_RATE_RPS=0` to disable. |
| `MAX_BODY_BYTES` | `65536` | Upper bound on JSON body size for orders. |

The server loads promo via `promo.LoadPromo`, which resolves shards from `PROMO_SHARDS_DIR`, then `./shards_seq`, then `<COUPON_DATA_DIR>/shards_seq`.

After building shards, `manifest.json` is emitted by `cmd/preprocessor_seq` (or regenerate with `go run ./cmd/genmanifest -dir ./shards_seq`). For production rollouts, point `PROMO_SHARDS_DIR` at a new directory atomically (symlink swap or config change) once the manifest validates.

## Building the shard index (offline)

From repo root, the one-shot script downloads the three corpora (if missing) and runs the preprocessor:

```bash
make setup-promo-shards
# or: ./scripts/setup-promo-shards.sh
```

`make fetch-coupons` (or `scripts/fetch-coupons.sh`) **only downloads** the three `couponbase*.gz` files — use it when you need the raw corpora without rebuilding shards.

Equivalent manual command:

```bash
go run ./cmd/preprocessor_seq \
  -data ./data \
  -out ./shards_seq \
  -tmp ./shards_seq/tmp \
  -scanLogEveryLines 5000000
```

- **`-out`:** final artifacts; deploy this directory (or an immutable versioned path).
- **`-tmp`:** scratch space; can be **many GB** during the run. Remove after success: `rm -rf ./shards_seq/tmp`.

**Rolling forward:** build to a new directory (e.g. `shards_seq_20260330/`), verify, then point `PROMO_SHARDS_DIR` at it and restart. Keeps rollback trivial.

## Running the server (local)

```bash
export PRODUCTS_PATH=./data/products.json
export COUPON_DATA_DIR=./data
export PROMO_SHARDS_DIR=./shards_seq
export API_KEY=apitest
go run ./cmd/server
```

Or use `make run` if your `Makefile` sets the same (adjust env as needed).

## Docker

See root [README](../README.md#docker). The **runtime image** is intentionally minimal:

- `server` binary  
- `data/products.json` only (not the full `data/` tree)  
- `shards_seq/*.bin` only (`.dockerignore` drops `couponbase*.gz` and `shards_seq/tmp/` from the build context)

Mount volumes if you prefer not to bake shards or products into the image. Set:

- `COUPON_DATA_DIR=/app/data` (or your mount)
- `PROMO_SHARDS_DIR=/app/shards_seq` (or mount path)

## Health and readiness

- `GET /health` — process up.
- For **readiness** with shards, optionally add a startup check in ops (not necessarily in code): confirm `PROMO_SHARDS_DIR` exists and `000.bin` is readable.

## CI

The repository CI runs `go test ./...` and build; it does **not** need the full gzip corpora or shards unless you add integration jobs.
