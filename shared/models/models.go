package models

import "time"

// User represents a user in the system
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    Password  string    `json:"-"` // Never expose in JSON
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Product represents a product in the catalog
type Product struct {
    ID            int64     `json:"id"`
    Name          string    `json:"name"`
    Description   string    `json:"description"`
    Price         float64   `json:"price"`
    SKU           string    `json:"sku"`
    CategoryID    *int64    `json:"category_id"`
    StockQuantity int       `json:"stock_quantity"`
    ImageURL      string    `json:"image_url"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

// Category represents a product category
type Category struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
    ID        int64     `json:"id"`
    CartID    string    `json:"cart_id"`
    ProductID int64     `json:"product_id"`
    Quantity  int       `json:"quantity"`
    Price     float64   `json:"price"` 
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Cart represents a shopping cart
type Cart struct {
    ID        string      `json:"id"`
    UserID    string      `json:"user_id"`
    Items     []CartItem  `json:"items"`
    Total     float64     `json:"total"`
    Status    string      `json:"status"` // active, checked_out, abandoned
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
}

// Order represents a customer order
type Order struct {
    ID        int64       `json:"id"`
    UserID    string      `json:"user_id"`
    Items     []OrderItem `json:"items"`
    Total     float64     `json:"total"`
    Status    string      `json:"status"` // pending, confirmed, shipped, delivered, cancelled
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
    ShippedAt *time.Time  `json:"shipped_at"`
    DeliveredAt *time.Time `json:"delivered_at"`
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

// SagaState tracks distributed transaction state
type SagaState struct {
    ID              string    `json:"id"`
    SagaType        string    `json:"saga_type"` // order_creation, cart_checkout, etc.
    Status          string    `json:"status"`    // pending, step1_complete, step2_complete, failed, compensated
    CorrelationID   string    `json:"correlation_id"`
    Payload         map[string]interface{} `json:"payload"`
    CompensationLog []string  `json:"compensation_log"` // Track which steps have been compensated
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// IdempotencyRecord tracks processed events to prevent duplicates
type IdempotencyRecord struct {
    ID        string    `json:"id"`
    EventID   string    `json:"event_id"`
    ServiceName string  `json:"service_name"`
    Action    string    `json:"action"`
    Result    string    `json:"result"` // success, failed
    CreatedAt time.Time `json:"created_at"`
}