package main

import (
    "flag"
    "fmt"
    "log"
    "path/filepath"
    "strings"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    dbURL := flag.String("db", "postgresql://prost_admin:prost_password@localhost:5432/prost?sslmode=disable", "Database URL")
    migrationsPath := flag.String("path", "./db/", "Path to migrations folder (without file://)")
    direction := flag.String("direction", "up", "Migration direction: up or down")
    steps := flag.Int("steps", 0, "Number of steps (0 = all)")

    flag.Parse()

    // Convert Windows path to forward slashes for migrate
    absPath, err := filepath.Abs(*migrationsPath)
    if err != nil {
        log.Fatalf("❌ Failed to get absolute path: %v", err)
    }
    absPath = strings.ReplaceAll(absPath, "\\", "/")
    sourceURL := fmt.Sprintf("file://%s", absPath)

    fmt.Println("========================================")
    fmt.Println("🔄 Running Database Migrations")
    fmt.Println("========================================")
    fmt.Printf("Source URL: %s\n", sourceURL)
    fmt.Printf("Database: %s\n", *dbURL)
    fmt.Printf("Direction: %s\n", *direction)
    fmt.Println()

    // Create migrate instance
    m, err := migrate.New(sourceURL, *dbURL)
    if err != nil {
        log.Fatalf("❌ Failed to create migrator: %v", err)
    }
    defer m.Close()

    var migrateErr error

    switch *direction {
	case "up":
        if *steps == 0 {
            migrateErr = m.Up()
        } else {
            migrateErr = m.Steps(*steps)
        }
    case "down":
        if *steps == 0 {
            migrateErr = m.Down()
        } else {
            migrateErr = m.Steps(-*steps)
        }
    }

    if migrateErr != nil && migrateErr != migrate.ErrNoChange {
        log.Fatalf("❌ Migration failed: %v", migrateErr)
    }

    if migrateErr == migrate.ErrNoChange {
        fmt.Println("✓ No migrations to run")
    } else {
        fmt.Println("✅ Migrations completed successfully!")
    }

    fmt.Println("========================================")
}