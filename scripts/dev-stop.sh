#!/bin/bash

# dev-stop.sh - Stop the entire Prost development environment
# Usage: ./scripts/dev-stop.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🛑 Stopping Prost Development Environment..."
echo ""

cd "$PROJECT_ROOT"

# Check if any containers are running
if ! docker-compose ps | grep -q "Up"; then
    echo "ℹ️  No containers are currently running"
    exit 0
fi

echo "Stopping containers..."
docker-compose down

echo ""
echo "✅ All services stopped"
echo ""
echo "💾 Note: Volumes are preserved for data persistence"
echo ""
echo "🧹 To also remove volumes and clean up:"
echo "   docker-compose down -v"
