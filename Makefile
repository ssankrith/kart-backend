.PHONY: run run-bin test build docker fetch-coupons setup-promo-shards

# Run via `go run` (same as developing with `go run ./cmd/server` from repo root).
run:
	go run ./cmd/server

# Build then run the binary. Invoke `make` from the repo root so cwd is correct
# for `data/`, `.env`, and default paths — otherwise you will not see promo logs
# (or the process exits earlier on missing products).
run-bin: build
	./bin/server

test:
	go test ./... -count=1 -race

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

docker:
	docker build -t kart-backend:local .

fetch-coupons:
	chmod +x scripts/fetch-coupons.sh && ./scripts/fetch-coupons.sh data

# Download coupon gzips + run preprocessor_seq to produce shards_seq/000.bin … 255.bin (long-running).
setup-promo-shards:
	chmod +x scripts/setup-promo-shards.sh && ./scripts/setup-promo-shards.sh
