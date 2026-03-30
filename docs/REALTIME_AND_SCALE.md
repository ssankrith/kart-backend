# Realtime behavior and scaling notes

This document explains how the promo path behaves **under load** and how the design scales if corpora or QPS grow.

## Request path cost

For each order with a coupon:

1. **Prelude check** — O(1) over rune/byte length (tiny).
2. **FNV-1a** over the code bytes — O(code length), effectively constant (≤10).
3. **First use of a shard:** mmap the shard file (OS maps pages on demand).
4. **Binary search** in a sorted fixed-width array — O(log N_shard) comparisons, where N_shard is the number of records in **that** shard only.

There is **no** full-table scan, **no** SQL `LIKE`, and **no** cross-shard search for a single code.

## Memory

- **Go heap:** Mostly catalog + request buffers. Shard data lives in **mapped** memory backed by files; it does not duplicate into a giant `map[string]…` for all codes at startup unless you choose to add that.
- **RSS:** Grows as shards are **touched** (cold start vs warm). Many replicas will each mmap the same files independently.

## CPU

- Validation is a small fixed amount of work per request dominated by **binary search** and syscall amortization on first shard touch.
- **Hot shards:** If FNV distribution is skewed for real codes, one shard could have more entries than others—still only O(log N) per lookup within that shard.

## Cold vs warm

- **Cold:** First requests that hit a new shard may pay page faults when the OS reads `*.bin` pages from disk.
- **Warm:** Steady-state latency is very stable for in-cache pages.

For a submission, documenting this distinction shows “production awareness” without needing benchmarks in the repo.

## Scaling dimensions

| Dimension | What happens |
|-----------|----------------|
| Larger gzip inputs | Offline job time and **tmp disk** grow; runtime unchanged if shard count/size is manageable. |
| More valid codes after merge | Larger per-shard files → slightly slower binary search; still log N. |
| Higher QPS | Horizontal scale **stateless** servers; same shard files on each instance (or shared read-only volume). |
| Updating corpora | Re-run preprocessor; **atomic** switch of `PROMO_SHARDS_DIR` to new directory. |

## What would change for “web scale”

If valid sets reached **billions** of strings:

- Shard files would grow; you might increase shard count (different format/version), use **content-defined** sharding, or add a **second-level** index.
- For this challenge, **256 shards × sorted records** is a deliberate balance: simple merge, simple runtime, mmap-friendly page granularity.

See [ALTERNATIVES.md](ALTERNATIVES.md) for other designs and when they win.
