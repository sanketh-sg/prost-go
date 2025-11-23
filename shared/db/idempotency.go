package db

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

// IdempotencyStore manages idempotency records to prevent duplicate processing
type IdempotencyStore struct {
    conn *Connection
}

// NewIdempotencyStore creates a new idempotency store
func NewIdempotencyStore(conn *Connection) *IdempotencyStore {
    return &IdempotencyStore{conn: conn}
}

// RecordProcessed records that an event has been processed
func (is *IdempotencyStore) RecordProcessed(ctx context.Context, eventID, serviceName, action, result string) error {
    query := `
        INSERT INTO $schema.idempotency_records (event_id, service_name, action, result, created_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (event_id, service_name) DO NOTHING
    `

    query = replaceSchema(query, is.conn.Schema)

    _, err := is.conn.ExecContext(ctx, query, eventID, serviceName, action, result, time.Now().UTC())
    if err != nil {
        return fmt.Errorf("failed to record idempotency: %w", err)
    }

    return nil
}

// IsProcessed checks if an event has already been processed
func (is *IdempotencyStore) IsProcessed(ctx context.Context, eventID, serviceName string) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 FROM $schema.idempotency_records 
            WHERE event_id = $1 AND service_name = $2
        )
    `

    query = replaceSchema(query, is.conn.Schema)

    var exists bool
    err := is.conn.QueryRowContext(ctx, query, eventID, serviceName).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check idempotency: %w", err)
    }

    return exists, nil
}

// GetRecord retrieves an idempotency record
func (is *IdempotencyStore) GetRecord(ctx context.Context, eventID, serviceName string) (map[string]interface{}, error) {
    query := `
        SELECT event_id, service_name, action, result, created_at
        FROM $schema.idempotency_records
        WHERE event_id = $1 AND service_name = $2
    `

    query = replaceSchema(query, is.conn.Schema)

    var record map[string]interface{}
    record = make(map[string]interface{})

    var eventID2, service, action, result string
    var createdAt time.Time

    err := is.conn.QueryRowContext(ctx, query, eventID, serviceName).Scan(&eventID2, &service, &action, &result, &createdAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get idempotency record: %w", err)
    }

    record["event_id"] = eventID2
    record["service_name"] = service
    record["action"] = action
    record["result"] = result
    record["created_at"] = createdAt

    return record, nil
}