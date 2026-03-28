# Build
FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

# Run
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/server /app/server
COPY --from=build /src/config.yaml /app/config.yaml
COPY --from=build /src/data /app/data
ENV CONFIG_PATH=/app/config.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
