# Alternative approaches and tradeoffs

Same problem: **substring presence in ≥2 of 3 corpora** for candidate codes of length 8–10, evaluated on newline-bounded lines.

## 1. In-memory hash map at startup (`map[string]uint8` bitmask)

**Idea:** Scan all gzips; for each window, OR a corpus bit into `map[key]`. Valid if `popcount(bits) ≥ 2`.

| Pros | Cons |
|------|------|
| O(1) average lookup at request time | Huge **RAM** and **startup time** if the *intermediate* window set is massive |
| Simple mental model | Go string keys + map overhead at scale |
| No offline artifact | Every deploy rescans gzip unless you serialize the map |

**When it fits:** Small qualifying sets or small corpora (challenge-sized “toy” data).

## 2. Full scan / regexp per request

**Idea:** For each coupon, search all three gzips for substring.

| Pros | Cons |
|------|------|
| Trivial to implement | **Unacceptable** latency and IO at scale |

**Verdict:** Useful only for debugging.

## 3. SQLite (or any RDBMS) with `LIKE '%code%'`

**Idea:** Load lines or codes; query with `LIKE`.

| Pros | Cons |
|------|------|
| Familiar ops | `LIKE '%x%'` without FTS is **table scan**-ish; hard to index for arbitrary substrings |
| Persistence | Wrong tool for substring-of-line over huge text at high QPS |

**Verdict:** Fine for analytics; poor default for this lookup pattern at scale.

## 4. Full-text search (Elasticsearch, SQLite FTS5, trigram indexes)

**Idea:** Index text for substring or token search.

| Pros | Cons |
|------|------|
| Powerful queries | Heavy ops stack for a static dataset |
| | Substring matching still has caveats vs enumerated windows |

**Verdict:** Overkill unless search requirements expand beyond this coupon rule.

## 5. CBIX-style per-corpus sorted binary files + mmap (nine files: 3 corpora × 3 widths)

**Idea:** Offline sort/dedup per corpus per width; runtime binary search each corpus for the candidate length.

| Pros | Cons |
|------|------|
| Low heap, predictable | More files, more binary searches per request |
| | Slightly more implementation surface |

**Verdict:** Strong; this repo previously explored similar ideas. Shards reduce to **one** search per request after FNV routing.

## 6. Sharded sorted files (this repo’s approach)

**Idea:** Offline enumerate windows → dedup/sort per corpus per shard → merge for ≥2 corpora → **256 sorted shard files**; runtime **FNV → one mmap’d shard → binary search**.

| Pros | Cons |
|------|------|
| One binary search per lookup | Must keep **FNV + layout** in sync between builder and server |
| Good RAM story at runtime | Offline job uses **disk + sort** (or pure-Go external sort if you eliminate `sort`) |
| Scales horizontally with read-only artifacts | Operational need to **version** shard dirs |

**Verdict:** Good balance for **static corpora + high read QPS** and a clear scale story.

## Summary table

| Approach | Build cost | Serve memory | Serve latency | Ops complexity |
|----------|------------|--------------|---------------|----------------|
| Heap map | High CPU/RAM | High if huge | O(1) avg | Low (no artifact) |
| CBIX 9-file | Medium | Low | O(log N) × 3 corpora | Medium |
| Sharded (this repo) | Medium-high (external sort) | Low | O(log N) in one shard | Medium |
| SQLite LIKE | Medium | DB-sized | Poor | Medium |
| FTS / search engine | High | Service + index | Good with tuning | High |

For a **one-off submission** with a **small final valid set**, a heap map might be enough. The sharded pipeline demonstrates **scale thinking**: separate offline compute, bounded runtime memory, and straightforward horizontal scaling of stateless servers.
