# Build stage: compile the Go app (CGO required for sqlite driver)
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
RUN go build -ldflags="-w -s" -o featherproxy .

# Runtime stage: minimal image with binary and static assets
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/featherproxy .
COPY --from=builder /build/internal/ui_server/static ./internal/ui_server/static
EXPOSE 4545
ENTRYPOINT ["./featherproxy"]
