package events

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sanketh-sg/prost/shared/models"
)

// BaseEvent is the common structure for all events
type BaseEvent struct {
	EventID       string    `json:"event_id"`       // Unique ID for idempotency
	EventType     string    `json:"event_type"`     // ProductCreated, OrderPlaced, etc.
	AggregateID   string    `json:"aggregate_id"`   // ID of the entity being changed
	AggregateType string    `json:"aggregate_type"` // product, order, cart
	Version       string    `json:"version"`        // Event schema version for evolution
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"` // Links related events in saga
}

func NewBaseEvent(eventType, aggregateID, aggregateType, correlationID string) BaseEvent {
	return BaseEvent{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       "1",
		Timestamp:     time.Now().UTC(),
		CorrelationID: correlationID,
	}
}

// ==================== Product Events ====================

// ProductCreatedEvent fired when a new product is created
type ProductCreatedEvent struct {
	BaseEvent
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
	CategoryID  *int64  `json:"category_id"`
	ImageURL    string  `json:"image_url"`
}

// ProductUpdatedEvent fired when product details change
type ProductUpdatedEvent struct {
	BaseEvent
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

// StockReservedEvent fired when inventory is reserved for an order
type StockReservedEvent struct {
	BaseEvent
	ProductID     int64  `json:"product_id"`
	Quantity      int    `json:"quantity"`
	OrderID       int64  `json:"order_id"`
	ReservationID string `json:"reservation_id"` // Link for compensation
}

// StockReleasedEvent fired when reserved inventory is released (compensation)
type StockReleasedEvent struct {
	BaseEvent
	ProductID     int64  `json:"product_id"`
	Quantity      int    `json:"quantity"`
	ReservationID string `json:"reservation_id"` // Reference to original reservation
	Reason        string `json:"reason"`         // order_cancelled, order_failed, etc.
}

// ==================== Cart Events ====================

// ItemAddedToCartEvent fired when item is added to cart
type ItemAddedToCartEvent struct {
	BaseEvent
	CartID    string  `json:"cart_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"` // Price snapshot
}

// ItemRemovedFromCartEvent fired when item is removed from cart
type ItemRemovedFromCartEvent struct {
	BaseEvent
	CartID    string  `json:"cart_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"` // Price at time of removal
}

// CartClearedEvent fired when cart is emptied
type CartClearedEvent struct {
	BaseEvent
	CartID string `json:"cart_id"`
	UserID string `json:"user_id"`
}

// CartCheckoutInitiatedEvent fired when checkout process begins (saga start)
type CartCheckoutInitiatedEvent struct {
	BaseEvent
	CartID string             `json:"cart_id"`
	UserID string             `json:"user_id"`
	Total  float64            `json:"total"`
	Items  []models.OrderItem `json:"items"`
}

// ==================== Order Events ====================
// OrderCreatedEvent fired when order is created in pending state (before inventory confirmation)
type OrderCreatedEvent struct {
	BaseEvent
	OrderID int64              `json:"order_id"`
	UserID  string             `json:"user_id"`
	Total   float64            `json:"total"`
	Items   []models.OrderItem `json:"items"`
}

// OrderPlacedEvent fired when an order is created (saga step 1)
type OrderPlacedEvent struct {
	BaseEvent
	OrderID int64              `json:"order_id"`
	UserID  string             `json:"user_id"`
	Total   float64            `json:"total"`
	Items   []models.OrderItem `json:"items"`
}

// OrderConfirmedEvent fired when payment/inventory confirmed (saga completion)
type OrderConfirmedEvent struct {
	BaseEvent
	OrderID int64 `json:"order_id"`
}

// OrderFailedEvent fired when order processing fails (saga failure)
type OrderFailedEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderCancelledEvent fired when order is cancelled (compensation trigger)
type OrderCancelledEvent struct {
	BaseEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderShippedEvent fired when order is shipped
type OrderShippedEvent struct {
	BaseEvent
	OrderID        int64     `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	ShippedAt      time.Time `json:"shipped_at"`
}

// ==================== User Events ====================

// UserRegisteredEvent fired when user creates account
type UserRegisteredEvent struct {
	BaseEvent
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// UserProfileUpdatedEvent fired when user updates profile
type UserProfileUpdatedEvent struct {
	BaseEvent
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// ==================== Utility Functions ====================

// MarshalEvent converts any event to JSON bytes
func MarshalEvent(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}

// UnmarshalEvent converts JSON bytes to an event interface
func UnmarshalEvent(data []byte, eventType string) (interface{}, error) {
	switch eventType {
	case "ProductCreated":
		var event ProductCreatedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "ProductUpdated":
		var event ProductUpdatedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "StockReserved":
		var event StockReservedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "StockReleased":
		var event StockReleasedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "ItemAddedToCart":
		var event ItemAddedToCartEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "CartCleared":
		var event CartClearedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "CartCheckoutInitiated":
		var event CartCheckoutInitiatedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "OrderPlaced":
		var event OrderPlacedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "OrderConfirmed":
		var event OrderConfirmedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "OrderFailed":
		var event OrderFailedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "OrderCancelled":
		var event OrderCancelledEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "OrderShipped":
		var event OrderShippedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "UserRegistered":
		var event UserRegisteredEvent
		err := json.Unmarshal(data, &event)
		return event, err
	case "UserProfileUpdated":
		var event UserProfileUpdatedEvent
		err := json.Unmarshal(data, &event)
		return event, err
	default:
		return nil, errors.New("unknown event type: " + eventType)
	}
}

//===================== Helper Functions for Idempotency Check ==============

// GetEventID returns the event ID from base event
func (be BaseEvent) GetEventID() string {
	return be.EventID
}

func (e ProductCreatedEvent) GetEventID() string {
	return e.EventID
}

func (e ProductUpdatedEvent) GetEventID() string {
	return e.EventID
}

func (e StockReservedEvent) GetEventID() string {
	return e.EventID
}

func (e StockReleasedEvent) GetEventID() string {
	return e.EventID
}

func (e ItemAddedToCartEvent) GetEventID() string {
	return e.EventID
}

func (e ItemRemovedFromCartEvent) GetEventID() string {
	return e.EventID
}

func (e CartClearedEvent) GetEventID() string {
	return e.EventID
}

func (e CartCheckoutInitiatedEvent) GetEventID() string {
	return e.EventID
}

func (e OrderCreatedEvent) GetEventID() string {
    return e.EventID
}

func (e OrderPlacedEvent) GetEventID() string {
	return e.EventID
}

func (e OrderConfirmedEvent) GetEventID() string {
	return e.EventID
}

func (e OrderFailedEvent) GetEventID() string {
	return e.EventID
}

func (e OrderCancelledEvent) GetEventID() string {
	return e.EventID
}

func (e OrderShippedEvent) GetEventID() string {
	return e.EventID
}

func (e UserRegisteredEvent) GetEventID() string {
	return e.EventID
}

func (e UserProfileUpdatedEvent) GetEventID() string {
	return e.EventID
}
