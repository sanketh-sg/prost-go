package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/products/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// InventoryReservationRepository handles inventory reservation database operations
type InventoryReservationRepository struct {
    conn *db.Connection
}

// NewInventoryReservationRepository creates new inventory repository
func NewInventoryReservationRepository(conn *db.Connection) *InventoryReservationRepository {
    return &InventoryReservationRepository{conn: conn}
}

// CreateReservation creates a new inventory reservation
func (ir *InventoryReservationRepository) CreateReservation(ctx context.Context, reservation *models.InventoryReservation) error {
    query := `
        INSERT INTO $schema.inventory_reservations 
        (product_id, quantity, order_id, reservation_id, status, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, product_id, quantity, order_id, reservation_id, status, created_at, expires_at
    `

    query = replaceSchema(query, ir.conn.Schema)

    err := ir.conn.QueryRowContext(ctx, query,
        reservation.ProductID,
        reservation.Quantity,
        reservation.OrderID,
        reservation.ReservationID,
        reservation.Status,
        reservation.CreatedAt,
        reservation.ExpiresAt,
    ).Scan(
        &reservation.ID,
        &reservation.ProductID,
        &reservation.Quantity,
        &reservation.OrderID,
        &reservation.ReservationID,
        &reservation.Status,
        &reservation.CreatedAt,
        &reservation.ExpiresAt,
    )

    if err != nil {
        log.Printf("Error creating reservation: %v", err)
        return fmt.Errorf("failed to create reservation: %w", err)
    }

    return nil
}

// GetReservation retrieves a reservation by reservation_id
func (ir *InventoryReservationRepository) GetReservation(ctx context.Context, reservationID string) (*models.InventoryReservation, error) {
    query := `
        SELECT id, product_id, quantity, order_id, reservation_id, status, created_at, expires_at, released_at
        FROM $schema.inventory_reservations
        WHERE reservation_id = $1
    `

    query = replaceSchema(query, ir.conn.Schema)

    reservation := &models.InventoryReservation{}
    err := ir.conn.QueryRowContext(ctx, query, reservationID).Scan(
        &reservation.ID,
        &reservation.ProductID,
        &reservation.Quantity,
        &reservation.OrderID,
        &reservation.ReservationID,
        &reservation.Status,
        &reservation.CreatedAt,
        &reservation.ExpiresAt,
        &reservation.ReleasedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get reservation: %w", err)
    }

    return reservation, nil
}

// GetReservationsByOrderID retrieves all reservations for an order
func (ir *InventoryReservationRepository) GetReservationsByOrderID(ctx context.Context, orderID int64) ([]*models.InventoryReservation, error) {
    query := `
        SELECT id, product_id, quantity, order_id, reservation_id, status, created_at, expires_at, released_at
        FROM $schema.inventory_reservations
        WHERE order_id = $1
    `

    query = replaceSchema(query, ir.conn.Schema)

    rows, err := ir.conn.QueryContext(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to get reservations: %w", err)
    }
    defer rows.Close()

    var reservations []*models.InventoryReservation
    for rows.Next() {
        reservation := &models.InventoryReservation{}
        err := rows.Scan(
            &reservation.ID,
            &reservation.ProductID,
            &reservation.Quantity,
            &reservation.OrderID,
            &reservation.ReservationID,
            &reservation.Status,
            &reservation.CreatedAt,
            &reservation.ExpiresAt,
            &reservation.ReleasedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan reservation: %w", err)
        }
        reservations = append(reservations, reservation)
    }

    return reservations, nil
}

// ReleaseReservation marks a reservation as released
func (ir *InventoryReservationRepository) ReleaseReservation(ctx context.Context, reservationID string) error {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = 'released', released_at = $1
        WHERE reservation_id = $2 AND status = 'reserved'
    `

    query = replaceSchema(query, ir.conn.Schema)

    result, err := ir.conn.ExecContext(ctx, query, time.Now().UTC(), reservationID)
    if err != nil {
        return fmt.Errorf("failed to release reservation: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("reservation not found or already released")
    }

    return nil
}

// ExpireReservations expires old reservations
func (ir *InventoryReservationRepository) ExpireReservations(ctx context.Context) (int64, error) {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = 'expired'
        WHERE status = 'reserved' AND expires_at < NOW()
    `

    query = replaceSchema(query, ir.conn.Schema)

    result, err := ir.conn.ExecContext(ctx, query)
    if err != nil {
        return 0, fmt.Errorf("failed to expire reservations: %w", err)
    }

    return result.RowsAffected()
}

// GetProductReservations gets active reservations for a product
func (ir *InventoryReservationRepository) GetProductReservations(ctx context.Context, productID int64) (int, error) {
    query := `
        SELECT COALESCE(SUM(quantity), 0)
        FROM $schema.inventory_reservations
        WHERE product_id = $1 AND status = 'reserved'
    `

    query = replaceSchema(query, ir.conn.Schema)

    var totalReserved int
    err := ir.conn.QueryRowContext(ctx, query, productID).Scan(&totalReserved)
    if err != nil {
        return 0, fmt.Errorf("failed to get product reservations: %w", err)
    }

    return totalReserved, nil
}

// UpdateReservationStatusByOrderID updates all reservations for an order to a new status
// Used when order is confirmed, failed, or cancelled
func (ir *InventoryReservationRepository) UpdateReservationStatusByOrderID(ctx context.Context, orderID string, status string) error {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = $1, updated_at = NOW()
        WHERE order_id::text = $2
    `

    query = replaceSchema(query, ir.conn.Schema)

    result, err := ir.conn.ExecContext(ctx, query, status, orderID)
    if err != nil {
        log.Printf("Error updating reservation status: %v", err)
        return fmt.Errorf("failed to update reservation status: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        log.Printf("⚠️  No reservations found for order %s", orderID)
        return nil // Don't fail if no reservations found (idempotency)
    }

    log.Printf("✓ Updated %d reservation(s) to status '%s' for order %s", rowsAffected, status, orderID)
    return nil
}

// UpdateReservationStatus updates a single reservation status by order ID (for confirmed orders)
// Used when we know the exact order ID as int64
func (ir *InventoryReservationRepository) UpdateReservationStatus(ctx context.Context, orderID int64, status string) error {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = $1, updated_at = NOW()
        WHERE order_id = $2
    `

    query = replaceSchema(query, ir.conn.Schema)

    result, err := ir.conn.ExecContext(ctx, query, status, orderID)
    if err != nil {
        log.Printf("Error updating reservation status: %v", err)
        return fmt.Errorf("failed to update reservation status: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        log.Printf("⚠️  No reservations found for order %d", orderID)
        return nil // Don't fail if no reservations found (idempotency)
    }

    log.Printf("✓ Updated %d reservation(s) to status '%s' for order %d", rowsAffected, status, orderID)
    return nil
}