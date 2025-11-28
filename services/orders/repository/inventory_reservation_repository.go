package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/orders/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// InventoryReservationRepository handles inventory reservation operations
type InventoryReservationRepository struct {
    conn *db.Connection
}

// NewInventoryReservationRepository creates new inventory reservation repository
func NewInventoryReservationRepository(conn *db.Connection) *InventoryReservationRepository {
    return &InventoryReservationRepository{conn: conn}
}

// CreateReservation creates new inventory reservation
func (irr *InventoryReservationRepository) CreateReservation(ctx context.Context, res *models.InventoryReservation) error {
    query := `
        INSERT INTO $schema.inventory_reservations 
        (id, order_id, product_id, quantity, reservation_id, status, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `

    query = replaceSchema(query, irr.conn.Schema)

    err := irr.conn.QueryRowContext(ctx, query,
        res.ID,
        res.OrderID,
        res.ProductID,
        res.Quantity,
        res.ReservationID,
        res.Status,
        res.CreatedAt,
        res.ExpiresAt,
    ).Scan(&res.ID)

    if err != nil {
        log.Printf("Error creating inventory reservation: %v", err)
        return fmt.Errorf("failed to create inventory reservation: %w", err)
    }

    return nil
}

// GetReservationsByOrderID retrieves all reservations for order
func (irr *InventoryReservationRepository) GetReservationsByOrderID(ctx context.Context, orderID int64) ([]*models.InventoryReservation, error) {
    query := `
        SELECT id, order_id, product_id, quantity, reservation_id, status, created_at, expires_at, released_at, fulfilled_at
        FROM $schema.inventory_reservations
        WHERE order_id = $1
    `

    query = replaceSchema(query, irr.conn.Schema)

    rows, err := irr.conn.QueryContext(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to get reservations: %w", err)
    }
    defer rows.Close()

    var reservations []*models.InventoryReservation
    for rows.Next() {
        res := &models.InventoryReservation{}
        err := rows.Scan(
            &res.ID,
            &res.OrderID,
            &res.ProductID,
            &res.Quantity,
            &res.ReservationID,
            &res.Status,
            &res.CreatedAt,
            &res.ExpiresAt,
            &res.ReleasedAt,
            &res.FulfilledAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan reservation: %w", err)
        }
        reservations = append(reservations, res)
    }

    return reservations, nil
}

// UpdateReservationStatus updates reservation status
func (irr *InventoryReservationRepository) UpdateReservationStatus(ctx context.Context, reservationID, status string) error {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = $1, fulfilled_at = CASE WHEN $1 = 'fulfilled' THEN NOW() ELSE NULL END
        WHERE reservation_id = $2
    `

    query = replaceSchema(query, irr.conn.Schema)

    _, err := irr.conn.ExecContext(ctx, query, status, reservationID)
    if err != nil {
        return fmt.Errorf("failed to update reservation status: %w", err)
    }

    return nil
}

// ReleaseReservation marks reservation as released
func (irr *InventoryReservationRepository) ReleaseReservation(ctx context.Context, reservationID string) error {
    query := `
        UPDATE $schema.inventory_reservations
        SET status = 'released', released_at = $1
        WHERE reservation_id = $2 AND status = 'reserved'
    `

    query = replaceSchema(query, irr.conn.Schema)

    result, err := irr.conn.ExecContext(ctx, query, time.Now().UTC(), reservationID)
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