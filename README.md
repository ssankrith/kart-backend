# kart-backend

Go implementation of the [Oolio advanced backend challenge](https://github.com/oolio-group/kart-challenge/blob/advanced-challenge/backend-challenge/README.md): food-ordering API per [OpenAPI 3.1](https://orderfoodonline.deno.dev/public/openapi.yaml), with **Gin**, **in-memory** product catalog (`data/products.json`), and **promo validation** using **precomputed shard files** (`000.bin`…`255.bin` under `PROMO_SHARDS_DIR`), built offline by `cmd/preprocessor_seq`.

## Features

- `GET /product` or `GET /api/product` — list products (same as [demo](https://orderfoodonline.deno.dev/api) `/api` prefix)  
- `GET /product/{productId}` — product by id (400 invalid id, 404 missing)  
- `POST /order` or `POST /api/order` — place order (requires `api_key` header; validates optional `couponCode` against Oolio rules)  
- `GET /health` — liveness  

**Promo rules:** UTF-8 length 8–10 inclusive; code must appear as a substring in **at least two** of the three corpora (`couponbase1.gz` … `couponbase3.gz` under `COUPON_DATA_DIR`, default `data/`). **Offline**, the preprocessor scans the gzips and emits only codes that satisfy the rule into shard files; **at runtime** the server looks up the candidate in the correct shard.

### Shipped `shards_seq` vs rebuilding the preprocessor

This repo **includes preprocessed promo shard files** in **`shards_seq/`** (`000.bin` … `255.bin`). That is enough to **`make run`**, run tests, and **build Docker** — you do **not** need the raw coupon gzips or a preprocessor run for normal use.

Run **`make setup-promo-shards`** (or `scripts/setup-promo-shards.sh`) **only if** you want to **verify, reproduce, or regenerate** those binaries (e.g. after changing corpora or preprocessor code). That step downloads `data/couponbase*.gz` and runs `cmd/preprocessor_seq` (long-running; large temp disk use). More detail at the top of [docs/README.md](docs/README.md).

## Documentation

- **[docs/README.md](docs/README.md)** — index to corpus stats, promo design (flowcharts), architecture, deployment, scaling notes, and alternatives (includes the note above on shipped shards).

## Quick start

```bash
make run          # or: go run ./cmd/server
make run-bin      # go build + ./bin/server (run from repo root for paths / .env)
```

```bash
curl -s localhost:8080/product | head -c 200
curl -s -H 'api_key: apitest' -H 'Content-Type: application/json' \
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHRS"}' \
  localhost:8080/order
```

## Environment (optional)

Values can be set in the process environment or in a **`.env`** file in the working directory (loaded on first access; does not override variables already set in the environment).

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `API_KEY` | `apitest` | Required value for header `api_key` on `POST /order` |
| `PRODUCTS_PATH` | `data/products.json` | Product catalog JSON |
| `COUPON_DATA_DIR` | `data` | Directory with `couponbase1.gz` … `3` |
| `PROMO_SHARDS_DIR` | (optional) | If set, preferred directory for `000.bin` … `255.bin`. Else `./shards_seq`, else `<COUPON_DATA_DIR>/shards_seq`. |

**Promo startup time (order of magnitude):** The server mmap-loads shard files. Startup is typically **seconds** (and shards are loaded lazily on first use).

**Request-time validation:** `Valid(code)` computes the shard via FNV-1a and binary-searches in the shard file (`11` bytes per sorted record). There is no SQLite or full-table `LIKE` in the server path.

**Code length vs corpora (fixed line lengths 8 / 9 / 10 per file):** A **9-character** code cannot appear as a substring of an **8-character** line, so corpus 1 contributes nothing for len-9 lookups (empty width-9 slice). A **10-character** code cannot fit in lines of length 8 or 9, so only corpus 3 can contain it; **at most one** corpus can match, so such codes can never satisfy “≥ two corpora” unless line lengths vary within files.

## Promo: optional — rebuild `shards_seq`

If you need to regenerate `shards_seq/*.bin` (not required for a fresh clone that already contains them):

```bash
make setup-promo-shards   # downloads corpora + runs cmd/preprocessor_seq (can take several minutes)
```

Or run `cmd/preprocessor_seq` yourself, e.g.:

```bash
go run ./cmd/preprocessor_seq -data ./data -out ./shards_seq -tmp ./shards_seq/tmp
```

The server resolves `PROMO_SHARDS_DIR` automatically (defaults try `./shards_seq`).

**Scripts:** **`make fetch-coupons`** (or `scripts/fetch-coupons.sh`) **only downloads** `data/couponbase1.gz` … `couponbase3.gz`. **`make setup-promo-shards`** does that download (if missing) **and** runs `cmd/preprocessor_seq` to rebuild `shards_seq/*.bin` — use it when you need the full offline pipeline, not when you only want the raw gzips.

## Docker

The image expects **`data/products.json`** and the committed **`shards_seq/`** shard files (`000.bin`…`255.bin`). You do **not** need to run the preprocessor before `make docker` unless you removed those bins. **`.dockerignore` excludes** the large `couponbase*.gz` files and `shards_seq/tmp/` so the image ships only the server binary, catalog JSON, and shard bins—not raw corpora or preprocessor scratch.

```bash
make docker
docker run --rm -p 8080:8080 \
  -e COUPON_DATA_DIR=/app/data \
  -e PROMO_SHARDS_DIR=/app/shards_seq \
  kart-backend:local
```

Override `PRODUCTS_PATH` / `PROMO_SHARDS_DIR` if you mount volumes instead of using the baked-in paths.

## CI

GitHub Actions (Go **1.24.x**, matching `go.mod`) runs `go test ./... -race`, `go build`, and `docker build` on push/PR — see [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

## License

MIT (or your choice) — challenge submission.
