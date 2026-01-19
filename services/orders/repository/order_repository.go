package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/orders/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// OrderRepository handles order database operations
type OrderRepository struct {
    conn *db.Connection
}

// NewOrderRepository creates new order repository
func NewOrderRepository(conn *db.Connection) *OrderRepository {
    return &OrderRepository{conn: conn}
}

// CreateOrder creates a new order
func (or *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
    query := `
        INSERT INTO $schema.orders 
        (id, user_id, cart_id, total, status, saga_correlation_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, user_id, cart_id, total, status, saga_correlation_id, created_at, updated_at
    `

    query = replaceSchema(query, or.conn.Schema)

    err := or.conn.QueryRowContext(ctx, query,
        order.ID,
        order.UserID,
        order.CartID,
        order.Total,
        order.Status,
        order.SagaCorrelationID,
        order.CreatedAt,
        order.UpdatedAt,
    ).Scan(
        &order.ID,
        &order.UserID,
        &order.CartID,
        &order.Total,
        &order.Status,
        &order.SagaCorrelationID,
        &order.CreatedAt,
        &order.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error creating order: %v", err)
        return fmt.Errorf("failed to create order: %w", err)
    }

    return nil
}

// GetOrder retrieves an order with items
func (or *OrderRepository) GetOrder(ctx context.Context, orderID int64) (*models.Order, error) {
    query := `
        SELECT id, user_id, cart_id, total, status, saga_correlation_id, 
               created_at, updated_at, shipped_at, delivered_at, cancelled_at
        FROM $schema.orders
        WHERE id = $1
    `

    query = replaceSchema(query, or.conn.Schema)

    order := &models.Order{}
    err := or.conn.QueryRowContext(ctx, query, orderID).Scan(
        &order.ID,
        &order.UserID,
        &order.CartID,
        &order.Total,
        &order.Status,
        &order.SagaCorrelationID,
        &order.CreatedAt,
        &order.UpdatedAt,
        &order.ShippedAt,
        &order.DeliveredAt,
        &order.CancelledAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get order: %w", err)
    }

    // Get order items
    itemsQuery := `
        SELECT id, order_id, product_id, quantity, price, created_at
        FROM $schema.order_items
        WHERE order_id = $1
        ORDER BY created_at ASC
    `

    itemsQuery = replaceSchema(itemsQuery, or.conn.Schema)

    rows, err := or.conn.QueryContext(ctx, itemsQuery, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to get order items: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        item := &models.OrderItem{}
        err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan order item: %w", err)
        }
        order.Items = append(order.Items, *item) //store the list of items in order
    }

    return order, nil
}

// GetOrdersByUserID retrieves all orders for a user
func (or *OrderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
    query := `
        SELECT id, user_id, cart_id, total, status, saga_correlation_id, 
               created_at, updated_at, shipped_at, delivered_at, cancelled_at
        FROM $schema.orders
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    query = replaceSchema(query, or.conn.Schema)

    rows, err := or.conn.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get orders by user: %w", err)
    }
    defer rows.Close()

    var orders []*models.Order
    for rows.Next() {
        order := &models.Order{}
        err := rows.Scan(
            &order.ID,
            &order.UserID,
            &order.CartID,
            &order.Total,
            &order.Status,
            &order.SagaCorrelationID,
            &order.CreatedAt,
            &order.UpdatedAt,
            &order.ShippedAt,
            &order.DeliveredAt,
            &order.CancelledAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        orders = append(orders, order)
    }

    return orders, nil
}

// AddOrderItem adds an item to an order
func (or *OrderRepository) AddOrderItem(ctx context.Context, item *models.OrderItem) error {
    query := `
        INSERT INTO $schema.order_items (order_id, product_id, quantity, price, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, order_id, product_id, quantity, price, created_at
    `

    query = replaceSchema(query, or.conn.Schema)

    err := or.conn.QueryRowContext(ctx, query,
        item.OrderID,
        item.ProductID,
        item.Quantity,
        item.Price,
        item.CreatedAt,
    ).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt)

    if err != nil {
        return fmt.Errorf("failed to add order item: %w", err)
    }

    return nil
}

// UpdateOrderStatus updates order status
func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
    query := `
        UPDATE $schema.orders
        SET status = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, or.conn.Schema)

    result, err := or.conn.ExecContext(ctx, query, status, time.Now().UTC(), orderID)
    if err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("order not found")
    }

    return nil
}

// CancelOrder cancels an order
func (or *OrderRepository) CancelOrder(ctx context.Context, orderID int64) error {
    query := `
        UPDATE $schema.orders
        SET status = 'cancelled', cancelled_at = $1, updated_at = $2
        WHERE id = $3 AND status != 'delivered'
    `

    query = replaceSchema(query, or.conn.Schema)

    result, err := or.conn.ExecContext(ctx, query, time.Now().UTC(), time.Now().UTC(), orderID)
    if err != nil {
        return fmt.Errorf("failed to cancel order: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("order not found or already delivered")
    }

    return nil
}

// Helper function
func replaceSchema(query, schema string) string {
    for i := 0; i < len(query)-len("$schema"); i++ {
        if query[i:i+len("$schema")] == "$schema" {
            query = query[:i] + schema + query[i+len("$schema"):]
        }
    }
    return query
}