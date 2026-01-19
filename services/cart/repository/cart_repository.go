package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/cart/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// CartRepository handles cart database operations
type CartRepository struct {
    conn *db.Connection
}

// NewCartRepository creates new cart repository
func NewCartRepository(conn *db.Connection) *CartRepository {
    return &CartRepository{conn: conn}
}

// CreateCart creates a new cart
func (cr *CartRepository) CreateCart(ctx context.Context, cart *models.Cart) error {
    query := `
        INSERT INTO $schema.carts (id, user_id, status, total, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, user_id, status, total, created_at, updated_at
    `

    query = replaceSchema(query, cr.conn.Schema)

    err := cr.conn.QueryRowContext(ctx, query,
        cart.ID,
        cart.UserID,
        cart.Status,
        cart.Total,
        cart.CreatedAt,
        cart.UpdatedAt,
    ).Scan(&cart.ID, &cart.UserID, &cart.Status, &cart.Total, &cart.CreatedAt, &cart.UpdatedAt)

    if err != nil {
        log.Printf("Error creating cart: %v", err)
        return fmt.Errorf("failed to create cart: %w", err)
    }

    return nil
}

// GetCart retrieves a cart with items
func (cr *CartRepository) GetCart(ctx context.Context, cartID string) (*models.Cart, error) {
    query := `
        SELECT id, user_id, status, total, created_at, updated_at, abandoned_at
        FROM $schema.carts
        WHERE id = $1 AND status != 'abandoned'
    `

    query = replaceSchema(query, cr.conn.Schema)

    cart := &models.Cart{}
    err := cr.conn.QueryRowContext(ctx, query, cartID).Scan(
        &cart.ID,
        &cart.UserID,
        &cart.Status,
        &cart.Total,
        &cart.CreatedAt,
        &cart.UpdatedAt,
        &cart.AbandonedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get cart: %w", err)
    }

    // Get cart items
    itemsQuery := `
        SELECT id, cart_id, product_id, quantity, price, created_at, updated_at
        FROM $schema.cart_items
        WHERE cart_id = $1
        ORDER BY created_at ASC
    `

    itemsQuery = replaceSchema(itemsQuery, cr.conn.Schema)

    rows, err := cr.conn.QueryContext(ctx, itemsQuery, cartID)
    if err != nil {
        return nil, fmt.Errorf("failed to get cart items: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        item := &models.CartItem{}
        err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan cart item: %w", err)
        }
        cart.Items = append(cart.Items, *item)
    }

    return cart, nil
}

// GetCartByUserID retrieves user's active cart
func (cr *CartRepository) GetCartByUserID(ctx context.Context, userID string) (*models.Cart, error) {
    query := `
        SELECT id, user_id, status, total, created_at, updated_at, abandoned_at
        FROM $schema.carts
        WHERE user_id = $1 AND status = 'active'
        ORDER BY created_at DESC
        LIMIT 1
    `

    query = replaceSchema(query, cr.conn.Schema)

    cart := &models.Cart{}
    err := cr.conn.QueryRowContext(ctx, query, userID).Scan(
        &cart.ID,
        &cart.UserID,
        &cart.Status,
        &cart.Total,
        &cart.CreatedAt,
        &cart.UpdatedAt,
        &cart.AbandonedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get cart by user id: %w", err)
    }

    // Get cart items
    itemsQuery := `
        SELECT id, cart_id, product_id, quantity, price, created_at, updated_at
        FROM $schema.cart_items
        WHERE cart_id = $1
        ORDER BY created_at ASC
    `

    itemsQuery = replaceSchema(itemsQuery, cr.conn.Schema)

    rows, err := cr.conn.QueryContext(ctx, itemsQuery, cart.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get cart items: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        item := &models.CartItem{}
        err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan cart item: %w", err)
        }
        cart.Items = append(cart.Items, *item)
    }

    return cart, nil
}

// AddItem adds an item to cart
func (cr *CartRepository) AddItem(ctx context.Context, item *models.CartItem) error {
    query := `
        INSERT INTO $schema.cart_items (id, cart_id, product_id, quantity, price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, cart_id, product_id, quantity, price, created_at, updated_at
    `

    query = replaceSchema(query, cr.conn.Schema)

    err := cr.conn.QueryRowContext(ctx, query,
        item.ID,
        item.CartID,
        item.ProductID,
        item.Quantity,
        item.Price,
        item.CreatedAt,
        item.UpdatedAt,
    ).Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to add item: %w", err)
    }

    return nil
}

// RemoveItem removes an item from cart
func (cr *CartRepository) RemoveItem(ctx context.Context, cartID string, productID int64) error {
    query := `
        DELETE FROM $schema.cart_items
        WHERE cart_id = $1 AND product_id = $2
    `

    query = replaceSchema(query, cr.conn.Schema)

    result, err := cr.conn.ExecContext(ctx, query, cartID, productID)
    if err != nil {
        return fmt.Errorf("failed to remove item: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("item not found in cart")
    }

    return nil
}

// UpdateCartStatus updates cart status
func (cr *CartRepository) UpdateCartStatus(ctx context.Context, cartID string, status string) error {
    query := `
        UPDATE $schema.carts
        SET status = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, cr.conn.Schema)

    result, err := cr.conn.ExecContext(ctx, query, status, time.Now().UTC(), cartID)
    if err != nil {
        return fmt.Errorf("failed to update cart status: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("cart not found")
    }

    return nil
}

// UpdateCartTotal updates cart total
func (cr *CartRepository) UpdateCartTotal(ctx context.Context, cartID string, total float64) error {
    query := `
        UPDATE $schema.carts
        SET total = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, cr.conn.Schema)

    _, err := cr.conn.ExecContext(ctx, query, total, time.Now().UTC(), cartID)
    if err != nil {
        return fmt.Errorf("failed to update cart total: %w", err)
    }

    return nil
}

// DeleteCart soft deletes a cart
func (cr *CartRepository) DeleteCart(ctx context.Context, cartID string) error {
    query := `
        UPDATE $schema.carts
        SET status = 'abandoned', abandoned_at = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, cr.conn.Schema)

    result, err := cr.conn.ExecContext(ctx, query, time.Now().UTC(), time.Now().UTC(), cartID)
    if err != nil {
        return fmt.Errorf("failed to delete cart: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("cart not found")
    }

    return nil
}

// ClearCart removes all items from cart
func (cr *CartRepository) ClearCart(ctx context.Context, cartID string) error {
    query := `DELETE FROM $schema.cart_items WHERE cart_id = $1`
    query = replaceSchema(query, cr.conn.Schema)

    _, err := cr.conn.ExecContext(ctx, query, cartID)
    if err != nil {
        return fmt.Errorf("failed to clear cart items: %w", err)
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