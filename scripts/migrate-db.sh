#!/bin/bash

set -e

echo "=== Prost Database Migration ==="

# Get project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check if PostgreSQL container is running
if ! docker ps | grep -q prost-postgres; then
    echo "✗ PostgreSQL container (prost-postgres) not running"
    echo "Run: docker-compose up -d postgres"
    exit 1
fi

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
for i in {1..30}; do
    if docker exec prost-postgres pg_isready -U prost_admin > /dev/null 2>&1; then
        echo "✓ PostgreSQL ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "✗ PostgreSQL timeout"
        exit 1
    fi
    sleep 1
done

# Create schemas
echo ""
echo "Creating schemas..."
docker exec prost-postgres psql -U prost_admin -d prost -c "
    CREATE SCHEMA IF NOT EXISTS users;
    CREATE SCHEMA IF NOT EXISTS catalog;
    CREATE SCHEMA IF NOT EXISTS cart;
    CREATE SCHEMA IF NOT EXISTS orders;
"
echo "✓ Schemas ready"

# Run migrations using goose on host
echo ""
echo "Running migrations..."
cd "$PROJECT_ROOT/infra/migrations"

# Check if goose installed
if ! command -v goose &> /dev/null; then
    echo "Installing goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
fi

# Connect to container's PostgreSQL from host
goose postgres "postgresql://prost_admin:prost_password@localhost:5432/prost?sslmode=disable" up

echo ""
echo "✓ Migrations complete!"
echo ""
echo "=== Verification ==="
docker exec prost-postgres psql -U prost_admin -d prost -c "
SELECT schemaname, COUNT(*) as tables
FROM pg_tables
WHERE schemaname IN ('users', 'catalog', 'cart', 'orders')
GROUP BY schemaname
ORDER BY schemaname;
"