.PHONY: run test build docker fetch-coupons

run:
	CONFIG_PATH=config.yaml go run ./cmd/server

test:
	go test ./... -count=1 -race

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

docker:
	docker build -t kart-backend:local .

fetch-coupons:
	chmod +x scripts/fetch-coupons.sh && ./scripts/fetch-coupons.sh data
