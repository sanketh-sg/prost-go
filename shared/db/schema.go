package db

import (
    "database/sql"
    "fmt"
)

// SchemaManager provides utilities for schema operations
type SchemaManager struct {
    db *sql.DB
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(db *sql.DB) *SchemaManager {
    return &SchemaManager{db: db}
}

// SetSearchPath sets the search path for a connection
// This ensures all queries default to the specified schema
func (sm *SchemaManager) SetSearchPath(schema string) error {
    query := fmt.Sprintf("SET search_path TO %s, public;", schema)
    _, err := sm.db.Exec(query)
    return err
}

// GetCurrentSchema returns the current schema in use
func (sm *SchemaManager) GetCurrentSchema() (string, error) {
    var currentSchema string
    err := sm.db.QueryRow("SELECT current_schema()").Scan(&currentSchema)
    return currentSchema, err
}

// TableExists checks if a table exists in the current schema
func (sm *SchemaManager) TableExists(tableName string) (bool, error) {
    var exists bool
    query := `
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_schema = current_schema() AND table_name = $1
        )
    `
    err := sm.db.QueryRow(query, tableName).Scan(&exists)
    return exists, err
}

// ListTables lists all tables in the current schema
func (sm *SchemaManager) ListTables() ([]string, error) {
    var tables []string
    query := `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = current_schema()
        ORDER BY table_name
    `
    rows, err := sm.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var tableName string
        if err := rows.Scan(&tableName); err != nil {
            return nil, err
        }
        tables = append(tables, tableName)
    }

    return tables, rows.Err()
}