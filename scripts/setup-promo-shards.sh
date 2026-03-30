#!/usr/bin/env bash
# Download challenge coupon corpora and build promo shard files (000.bin … 255.bin).
#
# Usage (from repo root):
#   chmod +x scripts/setup-promo-shards.sh && ./scripts/setup-promo-shards.sh
#
# Optional environment overrides:
#   DATA_DIR=./data              # where couponbase*..gz are stored (default: data)
#   OUT_DIR=./shards_seq         # output shard directory (default: shards_seq)
#   TMP_DIR=./shards_seq/tmp     # preprocessor scratch (default: OUT_DIR/tmp)
#   SCAN_LOG_EVERY=5000000       # progress log every N lines per corpus
#   SORT_MEM_MB=0                # optional: passed to OS sort (0 = default)
#   KEEP_TMP=1                   # if set, do not delete TMP_DIR after success
#
# Requires: curl, go 1.24+, sort (for preprocessor_seq), enough disk for tmp (often multi-GB).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DATA_DIR="${DATA_DIR:-data}"
OUT_DIR="${OUT_DIR:-shards_seq}"
TMP_DIR="${TMP_DIR:-$OUT_DIR/tmp}"
SCAN_LOG_EVERY="${SCAN_LOG_EVERY:-5000000}"
SORT_MEM_MB="${SORT_MEM_MB:-0}"
KEEP_TMP="${KEEP_TMP:-0}"

die() { echo "error: $*" >&2; exit 1; }

command -v curl >/dev/null || die "curl not found"
command -v go >/dev/null || die "go not found"

mkdir -p "$DATA_DIR"
BASE="https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com"

echo "==> Downloading couponbase1.gz, couponbase2.gz, couponbase3.gz into $DATA_DIR/"
for i in 1 2 3; do
  out="$DATA_DIR/couponbase${i}.gz"
  if [[ -f "$out" ]]; then
    echo "    (exists, skipping) $out"
  else
    echo "    downloading -> $out"
    curl -fsSL -o "$out" "$BASE/couponbase${i}.gz"
  fi
done

echo "==> Building shard index with cmd/preprocessor_seq (this can take several minutes) ..."
PRE_ARGS=(
  -data "$DATA_DIR"
  -out "$OUT_DIR"
  -tmp "$TMP_DIR"
  -scanLogEveryLines "$SCAN_LOG_EVERY"
)
if [[ "$SORT_MEM_MB" != "0" ]]; then
  PRE_ARGS+=( -sortMemMB "$SORT_MEM_MB" )
fi

go run ./cmd/preprocessor_seq "${PRE_ARGS[@]}"

if [[ "$KEEP_TMP" != "1" ]]; then
  echo "==> Removing preprocessor scratch: $TMP_DIR"
  rm -rf "$TMP_DIR"
else
  echo "==> KEEP_TMP=1: left $TMP_DIR on disk (add to .gitignore if needed)"
fi

echo ""
echo "Done."
echo "  Shards: $OUT_DIR/000.bin … 255.bin"
echo "  Run server with e.g.:"
echo "    export PROMO_SHARDS_DIR=\"$ROOT/$OUT_DIR\""
echo "    export COUPON_DATA_DIR=\"$ROOT/$DATA_DIR\""
echo "    export PRODUCTS_PATH=\"$ROOT/$DATA_DIR/products.json\""
echo "    go run ./cmd/server"
