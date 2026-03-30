# Build
FROM golang:1.24-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

# Run — only the server binary, product catalog JSON, and precomputed shard *.bin files.
# (No couponbase*.gz, no shards_seq/tmp — see .dockerignore.)
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/server /app/server
COPY --from=build /src/data/products.json /app/data/products.json
COPY --from=build /src/shards_seq/ /app/shards_seq/
ENV PRODUCTS_PATH=/app/data/products.json
ENV COUPON_DATA_DIR=/app/data
ENV PROMO_SHARDS_DIR=/app/shards_seq
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
