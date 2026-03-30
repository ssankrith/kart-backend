# Corpus data report (`couponbase1.gz` … `couponbase3.gz`)

This document records **measurable facts** about the three challenge corpora and the **implications** for promo validation. It is meant to support the design choice (offline shards + runtime lookup) and to explain why the **number of valid codes after filtering** is orders of magnitude smaller than “~300M lines.”

## File inventory

| File | Role | Compressed size (bytes) |
|------|------|---------------------------|
| `couponbase1.gz` | Corpus 1 | ~655,069,725 |
| `couponbase2.gz` | Corpus 2 | ~729,014,156 |
| `couponbase3.gz` | Corpus 3 | ~737,687,803 |

Paths are relative to `COUPON_DATA_DIR` (default `data/`).

## Gzip uncompressed size (ISIZE trailer)

Single-member gzip files expose an **ISIZE** field in the last 4 bytes (little-endian, uncompressed size modulo \(2^{32}\)). For these corpora the values are:

| File | ISIZE (unc. bytes, mod \(2^{32}\)) |
|------|-------------------------------------|
| `couponbase1.gz` | 965,346,993 |
| `couponbase2.gz` | 1,072,607,752 |
| `couponbase3.gz` | 1,084,227,656 |

## Line length structure (empirical)

**Method:** For each file, stream-decompress with the standard library gzip reader and measure **non-empty** line lengths after stripping `\r`/`\n`. **Sample size:** first 500,000 non-empty lines per file (identical results across the sample: one dominant length only).

| File | Lines in sample | Min length | Max length | Dominant length |
|------|-------------------|------------|------------|-----------------|
| `couponbase1.gz` | 500,000 | 8 | 8 | 8 (100% of sample) |
| `couponbase2.gz` | 500,000 | 9 | 9 | 9 (100% of sample) |
| `couponbase3.gz` | 500,000 | 10 | 10 | 10 (100% of sample) |

**Interpretation (strong evidence, not a formal proof of every line):** Each corpus appears to be **homogeneous in line length**: corpus 1 is all **8-byte** lines, corpus 2 all **9-byte**, corpus 3 all **10-byte** lines (plus newline). If that holds for the full files, the structure below follows.

## Approximate line counts

Assuming one `\n` per line (and no extra bytes), uncompressed size ≈ `lines × (line_length + 1)`:

| File | Assumed line bytes | Approx. lines |
|------|---------------------|---------------|
| `couponbase1.gz` | 8 | ISIZE / 9 ≈ **107,260,777** |
| `couponbase2.gz` | 9 | ISIZE / 10 ≈ **107,260,775** |
| `couponbase3.gz` | 10 | ISIZE / 11 ≈ **98,566,150** |

So “~300M lines” in conversation is in the right **order of magnitude** for **total lines across files**, not the count of *distinct valid promo strings* under the challenge rule.

## Why “≥2 of 3 corpora” yields a small final set here

**Rule (challenge):** A candidate code (UTF-8 length 8–10) is valid if it appears as a **substring** in **at least two** of the three corpora.

**Sliding windows:** For each non-empty line, we consider every contiguous byte substring of length 8, 9, and 10. That matches “substring of the line” for ASCII codes of those lengths.

**Combined with fixed line lengths per file:**

- **8-byte windows:** Possible in corpus 1 (line length 8): exactly **one** window per line (the whole line). Corpus 2 (length 9) contributes **two** overlapping 8-byte windows per line; corpus 3 (length 10) contributes **three** per line. So an 8-byte string can appear in more than one corpus.
- **9-byte windows:** Impossible in corpus 1 (lines are only 8 bytes). Only corpora 2 and 3 can contribute. So a 9-byte code can only ever match **at most two** corpora (2 and 3). Still valid if it appears in both.
- **10-byte windows:** Impossible in corpora 1 and 2. Only corpus 3 has length ≥ 10. So a 10-byte substring can appear in **at most one** corpus → **no 10-byte code can satisfy “≥2 corpora”** under this data shape.

Therefore the **final shard index should contain only 8- and 9-byte records** (and in practice mostly 8-byte), with **zero** length-10 entries. That matches a small, finite set driven by **overlap** between corpora, not raw line count.

## Relating to shard output size

After offline preprocessing (`cmd/preprocessor_seq`), the total number of emitted valid codes equals the sum over shards of `(file_size_bytes / 11)` for `000.bin` … `255.bin`. On a representative build this total was on the order of **tens of thousands** of records—consistent with the structural argument above.

## If data assumptions change

If future corpora have **mixed line lengths** within a file, or different encodings:

- The **same algorithm** still applies (windows + ≥2 corpora).
- The **cardinality** of the output can grow or shrink dramatically.
- Any claim like “only ~41k valid codes” must be **re-measured** on the new files.

Re-run corpus sampling and rebuild `shards_seq` when inputs change.
