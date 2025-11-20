#!/bin/bash

# db-migrate.sh - Run database migrations
# Usage: ./scripts/db-migrate.sh [up|down|status|create]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

DB_USER="prost_admin"
DB_PASSWORD="prost_password"
DB_NAME="prost"
DB_HOST="localhost"
DB_PORT="5432"

# Connection string
DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

COMMAND="${1:-up}"
MIGRATION_DIR="$PROJECT_ROOT/infra/migrations/catalog"

echo "🗄️  Database Migration Tool"
echo ""

# Check if goose is installed
if ! command -v goose &> /dev/null; then
    echo "📥 Installing goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
    echo "✅ goose installed"
fi

case "$COMMAND" in
    up)
        echo "⬆️  Running migrations up..."
        goose -dir "$MIGRATION_DIR" postgres "$DB_URL" up
        echo "✅ Migrations completed"
        ;;
    down)
        echo "⬇️  Rolling back last migration..."
        goose -dir "$MIGRATION_DIR" postgres "$DB_URL" down
        echo "✅ Rollback completed"
        ;;
    status)
        echo "📊 Migration status:"
        goose -dir "$MIGRATION_DIR" postgres "$DB_URL" status
        ;;
    create)
        if [ -z "$2" ]; then
            echo "❌ Error: Migration name required"
            echo "Usage: ./scripts/db-migrate.sh create <migration_name>"
            exit 1
        fi
        echo "📝 Creating migration: $2"
        goose -dir "$MIGRATION_DIR" postgres "$DB_URL" create "$2" sql
        echo "✅ Migration created in $MIGRATION_DIR"
        ;;
    *)
        echo "❌ Unknown command: $COMMAND"
        echo ""
        echo "Usage: ./scripts/db-migrate.sh [command]"
        echo ""
        echo "Commands:"
        echo "  up       - Run all pending migrations"
        echo "  down     - Rollback last migration"
        echo "  status   - Show migration status"
        echo "  create   - Create new migration (requires name argument)"
        echo ""
        echo "Examples:"
        echo "  ./scripts/db-migrate.sh up"
        echo "  ./scripts/db-migrate.sh down"
        echo "  ./scripts/db-migrate.sh status"
        echo "  ./scripts/db-migrate.sh create add_new_field"
        exit 1
        ;;
esac

echo ""
