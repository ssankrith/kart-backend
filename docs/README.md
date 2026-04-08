# Documentation index

This folder supplements the root [README](../README.md) with deeper material for reviewers.

## Hosted deployment

The API is deployed on [Render](https://render.com) at **https://kart-backend-18if.onrender.com**. The service is set to **redeploy automatically when changes are pushed to the `main` branch**.

## Shipped artifacts vs rebuilding the preprocessor

The repository **includes the preprocessed promo shard files** under **`shards_seq/`** (`000.bin` … `255.bin`). That is enough for **`go run ./cmd/server`**, tests, and the **Docker** image: you do **not** need to download the raw coupon gzips or run the offline pipeline just to run the API.

Run the preprocessor **only if** you want to **verify**, **reproduce**, or **regenerate** those binaries (e.g. after changing corpora or the build code). From the repo root:

- **`make setup-promo-shards`** — downloads `data/couponbase*.gz` (if missing) and runs `cmd/preprocessor_seq` (long-running; needs disk for temp files). See [DEPLOYMENT.md](DEPLOYMENT.md) and `scripts/setup-promo-shards.sh`.

| Document | Purpose |
|----------|---------|
| [CORPUS_DATA.md](CORPUS_DATA.md) | The three gzip corpora: sizes, measured line-length structure, and why the final valid-code set is small. |
| [PROMO_DESIGN.md](PROMO_DESIGN.md) | Promo validation rules, offline `preprocessor_seq` pipeline, shard file format, runtime lookup (FNV + mmap + binary search). Includes flowcharts. |
| [ARCHITECTURE.md](ARCHITECTURE.md) | HTTP API → services → promo checker; one diagram of the running service. |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Environment variables, building shards, shipping artifacts, Docker notes, rollback. |
| [REALTIME_AND_SCALE.md](REALTIME_AND_SCALE.md) | How this behaves as a live service: latency, memory, cold cache, scaling levers. |
| [BENCHMARKS.md](BENCHMARKS.md) | Go benchmarks and optional HTTP load (`hey`) methodology. |
| [ALTERNATIVES.md](ALTERNATIVES.md) | Other ways to solve the same problem (heap map, CBIX-style index, SQLite, etc.) and tradeoffs. |
| [adr/001-offline-sharded-index.md](adr/001-offline-sharded-index.md) | ADR: why offline shards + mmap + binary search. |

**Suggested reading order:** CORPUS_DATA → PROMO_DESIGN → ARCHITECTURE → DEPLOYMENT → REALTIME_AND_SCALE → ALTERNATIVES.
