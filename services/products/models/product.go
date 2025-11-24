package models

import (
    "time"

    "github.com/google/uuid"
)

// Category represents a product category
type Category struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
    ID            int64      `json:"id"`
    Name          string     `json:"name"`
    Description   string     `json:"description"`
    Price         float64    `json:"price"`
    SKU           string     `json:"sku"`
    CategoryID    *int64     `json:"category_id"`
    StockQuantity int        `json:"stock_quantity"`
    ImageURL      string     `json:"image_url"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
    DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// InventoryReservation tracks reserved inventory for orders
type InventoryReservation struct {
    ID            string     `json:"id"`
    ProductID     int64      `json:"product_id"`
    Quantity      int        `json:"quantity"`
    OrderID       int64      `json:"order_id"`
    ReservationID string     `json:"reservation_id"`
    Status        string     `json:"status"` // reserved, released, expired
    CreatedAt     time.Time  `json:"created_at"`
    ExpiresAt     time.Time  `json:"expires_at"`
    ReleasedAt    *time.Time `json:"released_at,omitempty"`
}

// CreateProductRequest request body for creating product
type CreateProductRequest struct {
    Name        string   `json:"name" binding:"required"`
    Description string   `json:"description"`
    Price       float64  `json:"price" binding:"required,gt=0"`
    SKU         string   `json:"sku" binding:"required"`
    CategoryID  *int64   `json:"category_id"`
    Stock       int      `json:"stock" binding:"required,gte=0"`
    ImageURL    string   `json:"image_url"`
}

// UpdateProductRequest request body for updating product
type UpdateProductRequest struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Price       float64  `json:"price"`
    Stock       int      `json:"stock"`
    ImageURL    string   `json:"image_url"`
}

// CreateCategoryRequest request body for creating category
type CreateCategoryRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

// ReserveInventoryRequest request to reserve inventory
type ReserveInventoryRequest struct {
    ProductID     int64  `json:"product_id" binding:"required"`
    Quantity      int    `json:"quantity" binding:"required,gt=0"`
    OrderID       int64  `json:"order_id" binding:"required"`
    ReservationID string `json:"reservation_id" binding:"required"`
}

// ReleaseInventoryRequest request to release reserved inventory
type ReleaseInventoryRequest struct {
    ReservationID string `json:"reservation_id" binding:"required"`
    Reason        string `json:"reason"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

// NewCategory creates new category
func NewCategory(name, description string) *Category {
    now := time.Now().UTC()
    return &Category{
        Name:        name,
        Description: description,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
}

// NewProduct creates new product
func NewProduct(name, description string, price float64, sku string, categoryID *int64, stock int, imageURL string) *Product {
    now := time.Now().UTC()
    return &Product{
        Name:          name,
        Description:   description,
        Price:         price,
        SKU:           sku,
        CategoryID:    categoryID,
        StockQuantity: stock,
        ImageURL:      imageURL,
        CreatedAt:     now,
        UpdatedAt:     now,
    }
}

// NewInventoryReservation creates new reservation
func NewInventoryReservation(productID int64, quantity int, orderID int64, reservationID string) *InventoryReservation {
    now := time.Now().UTC()
    return &InventoryReservation{
        ID:            uuid.New().String(),
        ProductID:     productID,
        Quantity:      quantity,
        OrderID:       orderID,
        ReservationID: reservationID,
        Status:        "reserved",
        CreatedAt:     now,
        ExpiresAt:     now.Add(24 * time.Hour),
    }
}