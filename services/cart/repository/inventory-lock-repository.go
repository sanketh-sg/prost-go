package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/cart/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// InventoryLockRepository handles inventory lock database operations
type InventoryLockRepository struct {
    conn *db.Connection
}

// NewInventoryLockRepository creates new inventory lock repository
func NewInventoryLockRepository(conn *db.Connection) *InventoryLockRepository {
    return &InventoryLockRepository{conn: conn}
}

// CreateLock creates a new inventory lock
func (ilr *InventoryLockRepository) CreateLock(ctx context.Context, lock *models.InventoryLock) error {
    query := `
        INSERT INTO $schema.inventory_locks 
        (id, cart_id, product_id, quantity, reservation_id, status, locked_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, cart_id, product_id, quantity, reservation_id, status, locked_at, expires_at
    `

    query = replaceSchema(query, ilr.conn.Schema)

    err := ilr.conn.QueryRowContext(ctx, query,
        lock.ID,
        lock.CartID,
        lock.ProductID,
        lock.Quantity,
        lock.ReservationID,
        lock.Status,
        lock.LockedAt,
        lock.ExpiresAt,
    ).Scan(
        &lock.ID,
        &lock.CartID,
        &lock.ProductID,
        &lock.Quantity,
        &lock.ReservationID,
        &lock.Status,
        &lock.LockedAt,
        &lock.ExpiresAt,
    )

    if err != nil {
        log.Printf("Error creating inventory lock: %v", err)
        return fmt.Errorf("failed to create inventory lock: %w", err)
    }

    return nil
}

// GetLocksByCartID retrieves all locks for a cart
func (ilr *InventoryLockRepository) GetLocksByCartID(ctx context.Context, cartID string) ([]*models.InventoryLock, error) {
    query := `
        SELECT id, cart_id, product_id, quantity, reservation_id, status, locked_at, expires_at, released_at
        FROM $schema.inventory_locks
        WHERE cart_id = $1 AND status = 'locked'
    `

    query = replaceSchema(query, ilr.conn.Schema)

    rows, err := ilr.conn.QueryContext(ctx, query, cartID)
    if err != nil {
        return nil, fmt.Errorf("failed to get locks: %w", err)
    }
    defer rows.Close()

    var locks []*models.InventoryLock
    for rows.Next() {
        lock := &models.InventoryLock{}
        err := rows.Scan(
            &lock.ID,
            &lock.CartID,
            &lock.ProductID,
            &lock.Quantity,
            &lock.ReservationID,
            &lock.Status,
            &lock.LockedAt,
            &lock.ExpiresAt,
            &lock.ReleasedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan lock: %w", err)
        }
        locks = append(locks, lock)
    }

    return locks, nil
}

// ReleaseLock marks a lock as released
func (ilr *InventoryLockRepository) ReleaseLock(ctx context.Context, reservationID string) error {
    query := `
        UPDATE $schema.inventory_locks
        SET status = 'released', released_at = $1
        WHERE reservation_id = $2 AND status = 'locked'
    `

    query = replaceSchema(query, ilr.conn.Schema)

    result, err := ilr.conn.ExecContext(ctx, query, time.Now().UTC(), reservationID)
    if err != nil {
        return fmt.Errorf("failed to release lock: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("lock not found or already released")
    }

    return nil
}

// ReleaseCartLocks releases all locks for a cart
func (ilr *InventoryLockRepository) ReleaseCartLocks(ctx context.Context, cartID string) error {
    query := `
        UPDATE $schema.inventory_locks
        SET status = 'released', released_at = $1
        WHERE cart_id = $2 AND status = 'locked'
    `

    query = replaceSchema(query, ilr.conn.Schema)

    _, err := ilr.conn.ExecContext(ctx, query, time.Now().UTC(), cartID)
    if err != nil {
        return fmt.Errorf("failed to release cart locks: %w", err)
    }

    return nil
}

// ExpireLocks expires old locks
func (ilr *InventoryLockRepository) ExpireLocks(ctx context.Context) (int64, error) {
    query := `
        UPDATE $schema.inventory_locks
        SET status = 'expired'
        WHERE status = 'locked' AND expires_at < NOW()
    `

    query = replaceSchema(query, ilr.conn.Schema)

    result, err := ilr.conn.ExecContext(ctx, query)
    if err != nil {
        return 0, fmt.Errorf("failed to expire locks: %w", err)
    }

    return result.RowsAffected()
}