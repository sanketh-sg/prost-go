FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy root go.mod and shared module
COPY go.mod .
COPY shared/ ./shared/

# Copy service files
COPY services/orders/ ./services/orders/

WORKDIR /app/services/orders

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o orders main.go

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/orders/orders .

EXPOSE 8082

CMD ["./orders"]