FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy root go.mod and shared module
COPY go.mod .
COPY shared/ ./shared/

# Copy service files
COPY services/products/ ./services/products/

WORKDIR /app/services/products

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o products main.go

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/products/products .

EXPOSE 8080

CMD ["./products"]