# Benchmarks

## Go micro-benchmarks

From the repo root:

```bash
go test ./internal/promo -bench=. -benchmem -count=5
```

Typical cases (see `internal/promo/bench_test.go`):

| Benchmark | What it measures |
|-----------|------------------|
| `BenchmarkValid_WarmHit` | Repeated `Valid` after the shard is mmap’d |
| `BenchmarkValid_ColdFirstLookup` | New checker + first lookup per iteration (mmap cold path) |
| `BenchmarkValid_PreludeReject` | Invalid length / prelude — no shard I/O |
| `BenchmarkValid_WarmMiss` | Well-formed code absent from shards |
| `BenchmarkValid_NaiveMap` | In-memory `map` membership (test baseline only) |

Record **machine model**, **Go version**, and **OS** next to any numbers you publish (LinkedIn, README).

## HTTP load (manual)

### 1. Disable rate limiting (required for latency numbers)

The server defaults to **`ORDER_RATE_RPS=100`** with a burst. A tool like `hey` with high concurrency will mostly get **`429 Too Many Requests`**, and printed percentiles mix fast rejections with real work — **not** comparable to “order latency.”

For throughput/latency of successful orders, start the process with:

```bash
export ORDER_RATE_RPS=0
```

(`0` or negative disables the limiter; the server logs `ORDER_RATE_RPS<=0: rate limiting disabled`.)

Optionally use **`GIN_MODE=release`** so debug logging does not skew timings:

```bash
export GIN_MODE=release
```

Example:

```bash
cd /path/to/kart-backend
ORDER_RATE_RPS=0 GIN_MODE=release go run ./cmd/server
# or: ORDER_RATE_RPS=0 GIN_MODE=release make run
```

### 2. Run `hey`

Install [hey](https://github.com/rakyll/hey) (or use `wrk`).

**Orders without coupon:**

```bash
hey -n 3000 -c 40 -m POST -H 'Content-Type: application/json' -H 'api_key: apitest' \
  -d '{"items":[{"productId":"1","quantity":1}]}' \
  http://127.0.0.1:8080/order
```

**Orders with coupon:**

```bash
hey -n 3000 -c 40 -m POST -H 'Content-Type: application/json' -H 'api_key: apitest' \
  -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"HAPPYHRS"}' \
  http://127.0.0.1:8080/order
```

Adjust `--n` and `-c` to your machine; keep **`Status code distribution`** in mind (see below).

### 3. Interpreting p50 / p90 / p95 / p99 on **200** responses

**`hey` does not split latency histograms by HTTP status.** All percentiles are over **every completed request**.

- **If `Status code distribution` shows only `[200]` and the count matches `-n`:** every sample is a successful order. The printed **latency distribution** (50%% / 90%% / 95%% / 99%% lines) **is** success-path latency. Use those as **p50 / p90 / p95 / p99** for the post (convert seconds to ms if you prefer).
- **If you see `[429]` or other codes:** the percentiles are **not** “successful order” latency unless you filter elsewhere. Either:
  - **Fix the setup** (e.g. `ORDER_RATE_RPS=0`, or lower `-c`/`-n`, or raise `ORDER_RATE_RPS` / `ORDER_RATE_BURST` only for the test), **or**
  - Use a load generator that supports **per-status** reporting (e.g. [Vegeta](https://github.com/tsenart/vegeta) with a custom reporter, or a small script that records time only when `status == 200`).

### 4. Example run (illustrative)

One local run with **`ORDER_RATE_RPS=0`**, **`GIN_MODE=release`**, **`hey -n 3000 -c 40`**, **all responses HTTP 200**:

| Scenario | p50 | p90 | p95 | p99 |
|----------|-----|-----|-----|-----|
| No coupon | 0.8 ms | 2.3 ms | 3.0 ms | 6.1 ms |
| Coupon `HAPPYHRS` | 0.8 ms | 2.4 ms | 3.0 ms | 4.9 ms |

Repeat on your hardware and paste your own `go` + OS + CPU line from `hey` (or `uname -a` / `go version`). Do not treat this table as a SLA; it is a **methodology** + **sample shape**.

### 5. Optional: benchmark **with** the rate limiter on

To characterize **429 behavior** under abuse, keep default `ORDER_RATE_RPS` and report:

- Counts of **200 vs 429**
- That latency percentiles are dominated by rejections unless you filter by status

## CI

The GitHub Actions workflow runs a short benchmark smoke (`-benchtime=200ms`) so large regressions in the promo path are visible without blocking on full HTTP load tests.
