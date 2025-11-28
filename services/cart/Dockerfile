FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy root go.mod and shared module
COPY go.mod .
COPY shared/ ./shared/

# Copy service files
COPY services/cart/ ./services/cart/

WORKDIR /app/services/cart

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o cart main.go

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/cart/cart .

EXPOSE 8081

CMD ["./cart"]