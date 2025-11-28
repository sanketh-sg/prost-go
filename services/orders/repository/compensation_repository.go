package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/sanketh-sg/prost/services/orders/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// CompensationLogRepository handles compensation log operations
type CompensationLogRepository struct {
    conn *db.Connection
}

// NewCompensationLogRepository creates new compensation log repository
func NewCompensationLogRepository(conn *db.Connection) *CompensationLogRepository {
    return &CompensationLogRepository{conn: conn}
}

// CreateCompensationLog creates new compensation log entry
func (clr *CompensationLogRepository) CreateCompensationLog(ctx context.Context, log *models.CompensationLog) error {
    payloadJSON, err := json.Marshal(log.CompensationPayload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    query := `
        INSERT INTO $schema.compensation_log 
        (id, order_id, saga_correlation_id, compensation_event, compensation_payload, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `

    query = replaceSchema(query, clr.conn.Schema)

    err = clr.conn.QueryRowContext(ctx, query,
        log.ID,
        log.OrderID,
        log.SagaCorrelationID,
        log.CompensationEvent,
        payloadJSON,
        log.Status,
        log.CreatedAt,
    ).Scan(&log.ID)

    if err != nil {
        fmt.Printf("Error creating compensation log: %v\n", err)
        return fmt.Errorf("failed to create compensation log: %w", err)
    }

    return nil
}

// GetCompensationLogsByOrderID retrieves compensation logs for order
func (clr *CompensationLogRepository) GetCompensationLogsByOrderID(ctx context.Context, orderID int64) ([]*models.CompensationLog, error) {
    query := `
        SELECT id, order_id, saga_correlation_id, compensation_event, compensation_payload, status, created_at, completed_at
        FROM $schema.compensation_log
        WHERE order_id = $1
        ORDER BY created_at ASC
    `

    query = replaceSchema(query, clr.conn.Schema)

    rows, err := clr.conn.QueryContext(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to get compensation logs: %w", err)
    }
    defer rows.Close()

    var logs []*models.CompensationLog
    for rows.Next() {
        log := &models.CompensationLog{}
        var payloadJSON []byte

        err := rows.Scan(
            &log.ID,
            &log.OrderID,
            &log.SagaCorrelationID,
            &log.CompensationEvent,
            &payloadJSON,
            &log.Status,
            &log.CreatedAt,
            &log.CompletedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan compensation log: %w", err)
        }

        err = json.Unmarshal(payloadJSON, &log.CompensationPayload)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
        }

        logs = append(logs, log)
    }

    return logs, nil
}

// UpdateCompensationStatus updates compensation log status
func (clr *CompensationLogRepository) UpdateCompensationStatus(ctx context.Context, logID, status string) error {
    query := `
        UPDATE $schema.compensation_log
        SET status = $1, completed_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, clr.conn.Schema)

    _, err := clr.conn.ExecContext(ctx, query, status, time.Now().UTC(), logID)
    if err != nil {
        return fmt.Errorf("failed to update compensation status: %w", err)
    }

    return nil
}