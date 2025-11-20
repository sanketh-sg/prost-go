#!/bin/bash

# dev-start.sh - Start the entire Prost development environment
# Usage: ./scripts/dev-start.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Starting Prost Development Environment..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker daemon is not running"
    exit 1
fi

echo "📦 Building services..."
cd "$PROJECT_ROOT"
docker-compose build

echo ""
echo "🔨 Starting containers..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 5

# Check if services are running
SERVICES=("prost-postgres" "prost-redis" "prost-gateway" "prost-catalog" "prost-cart" "prost-orders" "prost-frontend")

for service in "${SERVICES[@]}"; do
    if docker ps | grep -q "$service"; then
        echo "✅ $service is running"
    else
        echo "⚠️  $service failed to start"
    fi
done

echo ""
echo "🎉 Development environment started!"
echo ""
echo "📋 Available Services:"
echo "   Frontend:  http://localhost:3000"
echo "   Gateway:   http://localhost:80"
echo "   Catalog:   http://localhost:8080"
echo "   Cart:      http://localhost:8081"
echo "   Orders:    http://localhost:8082"
echo "   Postgres:  localhost:5432"
echo "   Redis:     localhost:6379"
echo ""
echo "💡 Tips:"
echo "   - View logs:  docker-compose logs -f <service>"
echo "   - Stop all:   ./scripts/dev-stop.sh"
echo "   - Migrate DB: ./scripts/db-migrate.sh up"
echo "   - Seed data:  ./scripts/seed-data.sh"
