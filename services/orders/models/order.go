package models

import (
    "time"

    "github.com/google/uuid"
)

// Order represents an order
type Order struct {
    ID                 int64      `json:"id"`
    UserID             string     `json:"user_id"`
    CartID             string     `json:"cart_id"`
    Items              []OrderItem `json:"items"`
    Total              float64    `json:"total"`
    Status             string     `json:"status"` // pending, confirmed, shipped, delivered, cancelled
    SagaCorrelationID  string     `json:"saga_correlation_id"`
    CreatedAt          time.Time  `json:"created_at"`
    UpdatedAt          time.Time  `json:"updated_at"`
    ShippedAt          *time.Time `json:"shipped_at,omitempty"`
    DeliveredAt        *time.Time `json:"delivered_at,omitempty"`
    CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
    ID        int64     `json:"id"`
    OrderID   int64     `json:"order_id"`
    ProductID int64     `json:"product_id"`
    Quantity  int       `json:"quantity"`
    Price     float64   `json:"price"` // Price at time of purchase
    CreatedAt time.Time `json:"created_at"`
}

// SagaState tracks order creation saga
type SagaState struct {
    ID               string                 `json:"id"`
    CorrelationID    string                 `json:"correlation_id"`
    SagaType         string                 `json:"saga_type"` // order_creation_saga
    Status           string                 `json:"status"`    // pending, cart_validated, inventory_reserved, payment_processed, confirmed, failed, compensated
    OrderID          *int64                 `json:"order_id"`
    CartID           string                 `json:"cart_id"`
    UserID           string                 `json:"user_id"`
    Payload          map[string]interface{} `json:"payload"`
    CompensationLog  []string               `json:"compensation_log"` // list of compensation actions
    CreatedAt        time.Time              `json:"created_at"`
    UpdatedAt        time.Time              `json:"updated_at"`
    ExpiresAt        time.Time              `json:"expires_at"`
}

// CompensationLog tracks compensating actions
type CompensationLog struct {
    ID                  string                 `json:"id"`
    OrderID             int64                  `json:"order_id"`
    SagaCorrelationID   string                 `json:"saga_correlation_id"`
    CompensationEvent   string                 `json:"compensation_event"` // StockReleased, CartCleared
    CompensationPayload map[string]interface{} `json:"compensation_payload"`
    Status              string                 `json:"status"` // pending, completed, failed
    CreatedAt           time.Time              `json:"created_at"`
    CompletedAt         *time.Time             `json:"completed_at,omitempty"`
}

// InventoryReservation tracks reserved inventory for order
type InventoryReservation struct {
    ID            string     `json:"id"`
    OrderID       int64      `json:"order_id"`
    ProductID     int64      `json:"product_id"`
    Quantity      int        `json:"quantity"`
    ReservationID string     `json:"reservation_id"`
    Status        string     `json:"status"` // reserved, released, expired, fulfilled
    CreatedAt     time.Time  `json:"created_at"`
    ExpiresAt     time.Time  `json:"expires_at"`
    ReleasedAt    *time.Time `json:"released_at,omitempty"`
    FulfilledAt   *time.Time `json:"fulfilled_at,omitempty"`
}

// CreateOrderRequest request to create order
type CreateOrderRequest struct {
    UserID   string `json:"user_id" binding:"required"`
    CartID   string `json:"cart_id" binding:"required"`
    OrderID  int64  `json:"order_id" binding:"required"`
    Total    float64 `json:"total" binding:"required,gt=0"`
    Items    []OrderItemRequest `json:"items" binding:"required,min=1"`
}

// OrderItemRequest represents item in create request
type OrderItemRequest struct {
    ProductID int64   `json:"product_id" binding:"required"`
    Quantity  int     `json:"quantity" binding:"required,gt=0"`
    Price     float64 `json:"price" binding:"required,gt=0"`
}

// CancelOrderRequest request to cancel order
type CancelOrderRequest struct {
    Reason string `json:"reason"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

// NewOrder creates new order
func NewOrder(userID, cartID string, orderID int64, total float64, sagaCorrelationID string) *Order {
    now := time.Now().UTC()
    return &Order{
        ID:                orderID,
        UserID:            userID,
        CartID:            cartID,
        Items:             []OrderItem{},
        Total:             total,
        Status:            "pending",
        SagaCorrelationID: sagaCorrelationID,
        CreatedAt:         now,
        UpdatedAt:         now,
    }
}

// NewOrderItem creates new order item
func NewOrderItem(orderID, productID int64, quantity int, price float64) *OrderItem {
    now := time.Now().UTC()
    return &OrderItem{
        OrderID:   orderID,
        ProductID: productID,
        Quantity:  quantity,
        Price:     price,
        CreatedAt: now,
    }
}

// NewSagaState creates new saga state for order creation
func NewSagaState(cartID, userID, correlationID string) *SagaState {
    now := time.Now().UTC()
    return &SagaState{
        ID:              uuid.New().String(),
        CorrelationID:   correlationID,
        SagaType:        "order_creation_saga",
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

// NewCompensationLog creates new compensation log
func NewCompensationLog(orderID int64, sagaCorrelationID, event string, payload map[string]interface{}) *CompensationLog {
    now := time.Now().UTC()
    return &CompensationLog{
        ID:                uuid.New().String(),
        OrderID:           orderID,
        SagaCorrelationID: sagaCorrelationID,
        CompensationEvent: event,
        CompensationPayload: payload,
        Status:            "pending",
        CreatedAt:         now,
    }
}

// NewInventoryReservation creates new inventory reservation
func NewInventoryReservation(orderID, productID int64, quantity int, reservationID string) *InventoryReservation {
    now := time.Now().UTC()
    return &InventoryReservation{
        ID:            uuid.New().String(),
        OrderID:       orderID,
        ProductID:     productID,
        Quantity:      quantity,
        ReservationID: reservationID,
        Status:        "reserved",
        CreatedAt:     now,
        ExpiresAt:     now.Add(24 * time.Hour),
    }
}