package models

import (
    "time"

    "github.com/google/uuid"
)

// Cart represents a shopping cart
type Cart struct {
    ID          string      `json:"id"`
    UserID      string      `json:"user_id"`
    Items       []CartItem  `json:"items"`
    Total       float64     `json:"total"`
    Status      string      `json:"status"` // active, checked_out, abandoned
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    AbandonedAt *time.Time  `json:"abandoned_at,omitempty"`
}

// CartItem represents an item in the cart
type CartItem struct {
    ID        string    `json:"id"`
    CartID    string    `json:"cart_id"`
    ProductID int64     `json:"product_id"`
    Quantity  int       `json:"quantity"`
    Price     float64   `json:"price"` // Price snapshot at time of adding
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// InventoryLock tracks reserved stock for cart items
type InventoryLock struct {
    ID            string     `json:"id"`
    CartID        string     `json:"cart_id"`
    ProductID     int64      `json:"product_id"`
    Quantity      int        `json:"quantity"`
    ReservationID string     `json:"reservation_id"`
    Status        string     `json:"status"` // locked, released, expired
    LockedAt      time.Time  `json:"locked_at"`
    ExpiresAt     time.Time  `json:"expires_at"`
    ReleasedAt    *time.Time `json:"released_at,omitempty"`
}

// SagaState tracks checkout saga progress
type SagaState struct {
    ID               string                 `json:"id"`
    CorrelationID    string                 `json:"correlation_id"`
    SagaType         string                 `json:"saga_type"` // checkout_saga
    Status           string                 `json:"status"`    // pending, inventory_validated, confirmed, failed, compensated
    CartID           string                 `json:"cart_id"`
    UserID           string                 `json:"user_id"`
    Payload          map[string]interface{} `json:"payload"`
    CompensationLog  []string               `json:"compensation_log"`
    CreatedAt        time.Time              `json:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at"`
    ExpiresAt        time.Time              `json:"expires_at"`
}

// CreateCartRequest request to create cart
type CreateCartRequest struct {
    UserID string `json:"user_id" binding:"required"`
}

// AddItemRequest request to add item to cart
type AddItemRequest struct {
    ProductID int64   `json:"product_id" binding:"required"`
    Quantity  int     `json:"quantity" binding:"required,gt=0"`
    Price     float64 `json:"price" binding:"required,gt=0"`
}

// RemoveItemRequest request to remove item from cart
type RemoveItemRequest struct {
    ProductID int64 `json:"product_id" binding:"required"`
}

// UpdateCartRequest request to update cart
type UpdateCartRequest struct {
    Status string `json:"status"`
}

// CheckoutRequest request to checkout cart
type CheckoutRequest struct {
    OrderID int64  `json:"order_id" binding:"required"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

// NewCart creates new cart
func NewCart(userID string) *Cart {
    now := time.Now().UTC()
    return &Cart{
        ID:        uuid.New().String(),
        UserID:    userID,
        Items:     []CartItem{},
        Total:     0.00,
        Status:    "active",
        CreatedAt: now,
        UpdatedAt: now,
    }
}

// NewCartItem creates new cart item
func NewCartItem(cartID string, productID int64, quantity int, price float64) *CartItem {
    now := time.Now().UTC()
    return &CartItem{
        ID:        uuid.New().String(),
        CartID:    cartID,
        ProductID: productID,
        Quantity:  quantity,
        Price:     price,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

// NewInventoryLock creates new inventory lock
func NewInventoryLock(cartID string, productID int64, quantity int, reservationID string) *InventoryLock {
    now := time.Now().UTC()
    return &InventoryLock{
        ID:            uuid.New().String(),
        CartID:        cartID,
        ProductID:     productID,
        Quantity:      quantity,
        ReservationID: reservationID,
        Status:        "locked",
        LockedAt:      now,
        ExpiresAt:     now.Add(1 * time.Hour),
    }
}

// NewSagaState creates new saga state
func NewSagaState(cartID, userID, correlationID string) *SagaState {
    now := time.Now().UTC()
    return &SagaState{
        ID:              uuid.New().String(),
        CorrelationID:   correlationID,
        SagaType:        "checkout_saga",
        Status:          "pending",
        CartID:          cartID,
        UserID:          userID,
        Payload:         make(map[string]interface{}),
        CompensationLog: []string{},
        CreatedAt:       now,
        UpdatedAt:       now,
        ExpiresAt:       now.Add(24 * time.Hour),
    }
}