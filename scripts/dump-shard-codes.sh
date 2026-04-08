#!/usr/bin/env bash
# Dump all promo codes from final shard bins to one line-oriented text file.
# Run from the repository root (same as other scripts).
#
# Usage:
#   ./scripts/dump-shard-codes.sh [extra args to dump_shard_codes]
# Example:
#   ./scripts/dump-shard-codes.sh -dir ./shards_seq -out ./all_promo_codes.txt

set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
exec go run ./cmd/dump_shard_codes "$@"
