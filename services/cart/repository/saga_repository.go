package repository

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/lib/pq"
    "github.com/sanketh-sg/prost/services/cart/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// SagaStateRepository handles saga state database operations
type SagaStateRepository struct {
    conn *db.Connection
}

// NewSagaStateRepository creates new saga repository
func NewSagaStateRepository(conn *db.Connection) *SagaStateRepository {
    return &SagaStateRepository{conn: conn}
}

// CreateSagaState creates a new saga state
func (sr *SagaStateRepository) CreateSagaState(ctx context.Context, saga *models.SagaState) error {
    payloadJSON, err := json.Marshal(saga.Payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    compensationLog := pq.Array(saga.CompensationLog)

    query := `
        INSERT INTO $schema.saga_states 
        (id, correlation_id, saga_type, status, cart_id, payload, compensation_log, created_at, updated_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, correlation_id, saga_type, status, cart_id, payload, compensation_log, created_at, updated_at, expires_at
    `

    query = replaceSchema(query, sr.conn.Schema)

    var payloadJSONResp []byte
    var compensationLogResp pq.StringArray

    err = sr.conn.QueryRowContext(ctx, query,
        saga.ID,
        saga.CorrelationID,
        saga.SagaType,
        saga.Status,
        saga.CartID,
        payloadJSON,
        compensationLog,
        saga.CreatedAt,
        saga.UpdatedAt,
        saga.ExpiresAt,
    ).Scan(
        &saga.ID,
        &saga.CorrelationID,
        &saga.SagaType,
        &saga.Status,
        &saga.CartID,
        &payloadJSONResp,
        &compensationLogResp,
        &saga.CreatedAt,
        &saga.UpdatedAt,
        &saga.ExpiresAt,
    )

    if err != nil {
        log.Printf("Error creating saga state: %v", err)
        return fmt.Errorf("failed to create saga state: %w", err)
    }

    return nil
}

// GetSagaState retrieves saga state by correlation ID
func (sr *SagaStateRepository) GetSagaState(ctx context.Context, correlationID string) (*models.SagaState, error) {
    query := `
        SELECT id, correlation_id, saga_type, status, cart_id, payload, compensation_log, created_at, updated_at, expires_at
        FROM $schema.saga_states
        WHERE correlation_id = $1
    `

    query = replaceSchema(query, sr.conn.Schema)

    saga := &models.SagaState{}
    var payloadJSON []byte
    var compensationLog pq.StringArray

    err := sr.conn.QueryRowContext(ctx, query, correlationID).Scan(
        &saga.ID,
        &saga.CorrelationID,
        &saga.SagaType,
        &saga.Status,
        &saga.CartID,
        &payloadJSON,
        &compensationLog,
        &saga.CreatedAt,
        &saga.UpdatedAt,
        &saga.ExpiresAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get saga state: %w", err)
    }

    // Unmarshal payload
    err = json.Unmarshal(payloadJSON, &saga.Payload)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
    }

    saga.CompensationLog = []string(compensationLog)

    return saga, nil
}

// UpdateSagaStatus updates saga status
func (sr *SagaStateRepository) UpdateSagaStatus(ctx context.Context, correlationID string, status string) error {
    query := `
        UPDATE $schema.saga_states
        SET status = $1, updated_at = $2
        WHERE correlation_id = $3
    `

    query = replaceSchema(query, sr.conn.Schema)

    result, err := sr.conn.ExecContext(ctx, query, status, time.Now().UTC(), correlationID)
    if err != nil {
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("saga state not found")
    }

    return nil
}

// AddCompensation adds compensation action to log
func (sr *SagaStateRepository) AddCompensation(ctx context.Context, correlationID string, compensation string) error {
    query := `
        UPDATE $schema.saga_states
        SET compensation_log = array_append(compensation_log, $1), updated_at = $2
        WHERE correlation_id = $3
    `

    query = replaceSchema(query, sr.conn.Schema)

    _, err := sr.conn.ExecContext(ctx, query, compensation, time.Now().UTC(), correlationID)
    if err != nil {
        return fmt.Errorf("failed to add compensation: %w", err)
    }

    return nil
}

// UpdateSagaPayload updates saga payload
func (sr *SagaStateRepository) UpdateSagaPayload(ctx context.Context, correlationID string, payload map[string]interface{}) error {
    payloadJSON, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    query := `
        UPDATE $schema.saga_states
        SET payload = $1, updated_at = $2
        WHERE correlation_id = $3
    `

    query = replaceSchema(query, sr.conn.Schema)

    _, err = sr.conn.ExecContext(ctx, query, payloadJSON, time.Now().UTC(), correlationID)
    if err != nil {
        return fmt.Errorf("failed to update saga payload: %w", err)
    }

    return nil
}