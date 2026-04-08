# ADR 001: Offline sharded index for promo validation

## Status

Accepted (challenge implementation).

## Context

Coupon validity is defined against three large newline-delimited corpora: a code is valid only if it appears as a substring (8–10 bytes, line boundaries respected) in **at least two** of the three files.

At HTTP request time, scanning corpora or running broad substring queries against raw text does not meet predictable latency or memory goals for a stateless API.

## Decision

1. **Offline** (`cmd/preprocessor_seq`): enumerate candidate windows, deduplicate per corpus, merge to enforce the “≥ two corpora” rule, and emit **256 FNV-sharded** binary files of fixed-width sorted records.
2. **Runtime** (`internal/promo`): route by `FNV-1a(code) % 256`, **mmap** the shard once per process per shard (lazy), **binary search** for an exact match.

A **`manifest.json`** (checksums and sizes) is written next to the shards so deployments can verify integrity; optional **`PROMO_SHARDS_STRICT=1`** fails startup if the manifest is missing or inconsistent.

## Consequences

- **Pros:** Bounded lookup work per request; no database hot path; easy horizontal scaling (read-only shard files per replica).
- **Cons:** Preprocessor run for corpus changes; operational need to ship or generate shard binaries + manifest; shard skew possible but still O(log N) per shard.

## Alternatives considered

See [ALTERNATIVES.md](../ALTERNATIVES.md) (SQLite `LIKE`, in-memory heap map, larger shard counts, etc.).
