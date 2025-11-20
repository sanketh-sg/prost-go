.PHONY: help migrate-up migrate-down migrate-status db-seed

# Database migration commands using goose
GOOSE_DRIVER := postgres
GOOSE_DBSTRING := "user=prost_admin password=prost_password dbname=prost sslmode=disable host=localhost port=5432"
GOOSE_MIGRATION_DIR := ./infra/migrations

help:
	@echo "Migration Commands:"
	@echo "  make migrate-up          - Run all pending migrations"
	@echo "  make migrate-down        - Rollback last migration"
	@echo "  make migrate-status      - Show migration status"
	@echo "  make migrate-create      - Create new migration (NAME=filename)"
	@echo "  make db-seed             - Seed database with sample data"

# Install goose if not present
install-goose:
	@command -v goose >/dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest

# Run all pending migrations
migrate-up: install-goose
	@echo "Running migrations..."
	@goose -dir $(GOOSE_MIGRATION_DIR) postgres $(GOOSE_DBSTRING) up

# Rollback last migration
migrate-down: install-goose
	@echo "Rolling back last migration..."
	@goose -dir $(GOOSE_MIGRATION_DIR) postgres $(GOOSE_DBSTRING) down

# Show migration status
migrate-status: install-goose
	@echo "Migration status:"
	@goose -dir $(GOOSE_MIGRATION_DIR) postgres $(GOOSE_DBSTRING) status

# Create new migration
migrate-create: install-goose
	@goose -dir $(GOOSE_MIGRATION_DIR) postgres $(GOOSE_DBSTRING) create $(NAME) sql

# Docker-based migrations (runs inside container)
docker-migrate-up:
	@docker-compose exec catalog migrate -path /migrations -database "$(GOOSE_DBSTRING)" up

docker-migrate-down:
	@docker-compose exec catalog migrate -path /migrations -database "$(GOOSE_DBSTRING)" down

docker-migrate-status:
	@docker-compose exec catalog migrate -path /migrations -database "$(GOOSE_DBSTRING)" version
