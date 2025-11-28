# Multi-stage build for gateway service
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY gateway/go.mod gateway/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY gateway/ .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gateway .

# Runtime image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/gateway .

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/health || exit 1

EXPOSE 80

CMD ["./gateway"]