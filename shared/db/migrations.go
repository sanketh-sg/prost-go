package db

import (
    "database/sql"
    "fmt"
    "log"

    "github.com/pressly/goose/v3"
)

// Migration holds migration information
type Migration struct {
    Name string
    SQL  string
}

// RunMigrations runs all pending migrations for a given schema
func RunMigrations(db *sql.DB, schema string, migrations []Migration) error {
    log.Printf("Running migrations for schema: %s", schema)

    // Enable goose to use the same connection
    goose.SetBaseFS(nil)

    for _, migration := range migrations {
        // Create versioned migration SQL with schema switching
        migrationSQL := fmt.Sprintf("SET search_path TO %s; %s", schema, migration.SQL)

        _, err := db.Exec(migrationSQL)
        if err != nil {
            return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
        }

        log.Printf("✓ Migration applied: %s", migration.Name)
    }

    return nil
}

// CreateSchema creates a new schema if it doesn't exist
func CreateSchema(db *sql.DB, schemaName string) error {
    query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", schemaName)
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
    }

    log.Printf("✓ Schema created/verified: %s", schemaName)
    return nil
}

// DropSchema drops a schema (use with caution!)
func DropSchema(db *sql.DB, schemaName string) error {
    query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName)
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
    }

    log.Printf("✓ Schema dropped: %s", schemaName)
    return nil
}