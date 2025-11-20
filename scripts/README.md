# Development Scripts

Utility scripts for managing the Prost development environment.

## Scripts Overview

### `dev-start.sh`
Starts the entire development environment with Docker Compose.

**Usage:**
```bash
./scripts/dev-start.sh
```

**What it does:**
- Checks Docker is running
- Builds all services
- Starts containers
- Waits for services to be healthy
- Displays available endpoints

**Output:**
```
Frontend:  http://localhost:3000
Gateway:   http://localhost:80
Catalog:   http://localhost:8080
Cart:      http://localhost:8081
Orders:    http://localhost:8082
```

---

### `dev-stop.sh`
Stops all running containers without removing volumes (data is preserved).

**Usage:**
```bash
./scripts/dev-stop.sh
```

**What it does:**
- Stops all containers
- Preserves volumes for data persistence
- Shows how to completely clean up

**To remove volumes:**
```bash
docker-compose down -v
```

---

### `db-migrate.sh`
Manages database migrations using Goose.

**Usage:**
```bash
./scripts/db-migrate.sh [command]
```

**Commands:**

| Command | Description |
|---------|-------------|
| `up` | Run all pending migrations |
| `down` | Rollback last migration |
| `status` | Show migration status |
| `create <name>` | Create new migration file |

**Examples:**
```bash
# Run all migrations
./scripts/db-migrate.sh up

# Check status
./scripts/db-migrate.sh status

# Rollback last migration
./scripts/db-migrate.sh down

# Create new migration
./scripts/db-migrate.sh create add_discount_column
```

**Requirements:**
- PostgreSQL running and accessible on `localhost:5432`
- Goose installed (auto-installed on first run)
- Database: `prost` with user `prost_admin` / password `prost_password`

---

### `seed-data.sh`
Seeds the database with initial product and category data.

**Usage:**
```bash
./scripts/seed-data.sh
```

**What it does:**
- Connects to PostgreSQL database
- Inserts 5 categories (Beer, Wine, Spirits, Non-Alcoholic, Energy Drinks)
- Inserts 10 sample products with stock quantities
- Shows summary of seeded data
- Uses `ON CONFLICT DO NOTHING` to avoid duplicates on re-runs

**Sample Data:**
- **Categories:** Beer, Wine, Spirits, Non-Alcoholic, Energy Drinks
- **Products:** 10 diverse beverage products with realistic pricing and stock levels

**Requirements:**
- PostgreSQL client (`psql`) installed
- Database running and accessible
- Migrations already run (`./scripts/db-migrate.sh up`)

---

## Quick Start Workflow

Get the entire environment running in minutes:

```bash
# 1. Start the environment
./scripts/dev-start.sh

# 2. Wait for all containers to be healthy (should show ~15 seconds)

# 3. Run database migrations
./scripts/db-migrate.sh up

# 4. Seed initial data
./scripts/seed-data.sh

# 5. Access the application
# Frontend: http://localhost:3000
# API: http://localhost/api/
```

## Troubleshooting

### Docker daemon not running
```bash
# Start Docker Desktop (macOS/Windows) or docker daemon (Linux)
docker info
```

### Database connection error
```bash
# Make sure containers are running
docker-compose ps

# Check logs
docker-compose logs postgres
```

### psql not found
```bash
# macOS with Homebrew
brew install postgresql

# Ubuntu/Debian
sudo apt-get install postgresql-client

# Windows - Install PostgreSQL or use Windows Subsystem for Linux
```

### Migration permission issues
```bash
# Make scripts executable
chmod +x ./scripts/*.sh
```

## Environment Variables

All scripts use these default credentials:
- **User:** `prost_admin`
- **Password:** `prost_password`
- **Database:** `prost`
- **Host:** `localhost`
- **Port:** `5432`

To change these, edit the scripts or update `docker-compose.yml`.

## Monitoring

### View logs for all services
```bash
docker-compose logs -f
```

### View logs for specific service
```bash
docker-compose logs -f catalog
docker-compose logs -f gateway
docker-compose logs -f frontend
```

### Check container status
```bash
docker-compose ps
```

### Execute commands in container
```bash
docker-compose exec catalog /bin/sh
docker-compose exec postgres psql -U prost_admin -d prost
```

## Cleanup

### Stop containers (keep volumes)
```bash
./scripts/dev-stop.sh
```

### Stop and remove everything
```bash
docker-compose down -v
```

### Remove specific service
```bash
docker-compose rm -f catalog
```
