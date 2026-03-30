#!/usr/bin/env bash
set -euo pipefail
DIR="${1:-data}"
mkdir -p "$DIR"
base="https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com"
for i in 1 2 3; do
  echo "Downloading couponbase${i}.gz ..."
  curl -fsSL -o "$DIR/couponbase${i}.gz" "$base/couponbase${i}.gz"
done
echo "Done. Use COUPON_DATA_DIR=$DIR (or copy gzips into data/)"
