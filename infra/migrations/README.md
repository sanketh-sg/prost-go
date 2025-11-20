# Database Migrations

This directory contains database migration files for the Prost microservices.

## Directory Structure

```
infra/migrations/
├── catalog/          # Catalog service migrations
├── cart/             # Cart service migrations (future)
├── orders/           # Orders service migrations (future)
└── shared/           # Shared/common migrations (future)
```

## Migration Tool: Goose

We use **Goose** for database migrations - a simple and reliable migration tool for Go.

### Installation

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Usage

#### Run all pending migrations
```bash
make migrate-up
```

#### Rollback last migration
```bash
make migrate-down
```

#### Check migration status
```bash
make migrate-status
```

#### Create new migration
```bash
make migrate-create NAME=add_column_to_products
```

This creates a file like `002_add_column_to_products.sql` in the current migration directory.

## Migration File Format

Migration files use Goose SQL format:

```sql
-- +goose Up
-- Write your migration up here

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- +goose Down
-- Write your rollback here

DROP TABLE products;
```

## Current Migrations

### Catalog Service (`catalog/`)
- `001_create_catalog_tables.sql` - Initial schema with products and categories tables

## Best Practices

1. **Always include both Up and Down** - Ensure migrations are reversible
2. **Use transactions** - Wrap DDL in `-- +goose StatementBegin` and `-- +goose StatementEnd`
3. **One change per file** - Keep migrations focused and atomic
4. **Meaningful names** - Use descriptive filenames (e.g., `add_email_to_users`)
5. **Test rollbacks** - Always test `down` migrations before committing
6. **Schema versioning** - Goose maintains a `goose_db_version` table automatically

## Docker-based Migrations

When running migrations inside Docker:

```bash
docker-compose exec catalog goose -dir /migrations postgres "$DATABASE_URL" up
```

This is useful for CI/CD pipelines where the database is only accessible from within containers.

## Connection String

Default connection string (must match docker-compose.yml):

```
postgresql://prost_admin:prost_password@localhost:5432/prost
```

For Docker containers, use:

```
postgresql://prost_admin:prost_password@postgres:5432/prost
```
