package db

import (
    "context"
    "database/sql" // Standard SQL package
    "fmt"
    "log"
    "time"

    _ "github.com/lib/pq" // Postgres driver
)

// Config holds database configuration
type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    Schema   string
    SSLMode  string
}

// Connection holds the database connection pool
type Connection struct {
    DB     *sql.DB
    Schema string
}

// Initalize new database connection

func NewDBConnection(cfg Config) (*Connection, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	// Prevents connection failures when not set, PostgreSQL requires an SSL mode; empty value can cause connection refusal.

	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,)

	dbConn, err := sql.Open("postgres", dataSourceName)

	if err != nil {
		return nil, err
	}

	// Configure connection pool
    dbConn.SetMaxOpenConns(25)
    dbConn.SetMaxIdleConns(5)
    dbConn.SetConnMaxLifetime(5 * time.Minute)
    dbConn.SetConnMaxIdleTime(10 * time.Minute)

	// Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := dbConn.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Printf("Connected to PostgreSQL database: %s (schema: %s)", cfg.DBName, cfg.Schema)

    return &Connection{
        DB:     dbConn,
        Schema: cfg.Schema,
    }, nil
}


// Helper functions

func (c *Connection) DBConnClose() error {
    return c.DB.Close()
}


func contains(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}

func replaceSchema(query, schema string) string {
    // Simple replacement of $schema with actual schema name
    for {
        idx := 0
        for i := 0; i < len(query)-len("$schema"); i++ {
            if query[i:i+len("$schema")] == "$schema" {
                idx = i
                break
            }
        }
        if idx == 0 && query[:7] != "$schema" {
            break
        }
        query = query[:idx] + schema + query[idx+len("$schema"):]
    }
    return query
}


// PrepareStmt prepares a statement with schema substitution
// Usage: db.PrepareStmt(ctx, "SELECT * FROM $1.users WHERE id = $2")
// The $1 will be replaced with the schema name
func (c *Connection) PrepareStmt(ctx context.Context, query string) (*sql.Stmt, error) {
    // Replace schema placeholder if exists
    if schemaPlaceholder := "$schema"; contains(query, schemaPlaceholder) {
        query = replaceSchema(query, c.Schema)
    }

    stmt, err := c.DB.PrepareContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare statement: %w", err)
    }

    return stmt, nil
}

// QueryRowContext executes a query that returns a single row
func (c *Connection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return c.DB.QueryRowContext(ctx, query, args...)
}

// QueryContext executes a query that returns multiple rows
func (c *Connection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return c.DB.QueryContext(ctx, query, args...)
}

// ExecContext executes a query that doesn't return rows
func (c *Connection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return c.DB.ExecContext(ctx, query, args...)
}

// BeginTx starts a new transaction
func (c *Connection) BeginTx(ctx context.Context) (*sql.Tx, error) {
    return c.DB.BeginTx(ctx, nil)
}

