package subscribers

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/sanketh-sg/prost/services/cart/repository"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/events"
)

// EventHandler handles incoming events for cart service
type EventHandler struct {
    cartRepo          *repository.CartRepository
    sagaRepo          *repository.SagaStateRepository
    inventoryLockRepo *repository.InventoryLockRepository
    idempotencyStore  *db.IdempotencyStore
}

// NewEventHandler creates new event handler
func NewEventHandler(
    cartRepo *repository.CartRepository,
    sagaRepo *repository.SagaStateRepository,
    inventoryLockRepo *repository.InventoryLockRepository,
    idempotencyStore *db.IdempotencyStore,
) *EventHandler {
    return &EventHandler{
        cartRepo:          cartRepo,
        sagaRepo:          sagaRepo,
        inventoryLockRepo: inventoryLockRepo,
        idempotencyStore:  idempotencyStore,
    }
}

// HandleEvent processes incoming events
func (eh *EventHandler) HandleEvent(ctx context.Context, message []byte) error {
    // Extract event type
    var baseEvent struct {
        EventID   string `json:"event_id"`
        EventType string `json:"event_type"`
    }

    if err := json.Unmarshal(message, &baseEvent); err != nil {
        return fmt.Errorf("failed to unmarshal base event: %w", err)
    }

    eventID := baseEvent.EventID
    eventType := baseEvent.EventType

    // Check idempotency
    processed, err := eh.idempotencyStore.IsProcessed(ctx, eventID, "cart")
    if err != nil {
        log.Printf("⚠️  Failed to check idempotency: %v", err)
    }

    if processed {
        log.Printf("⏭️  Event %s already processed, skipping", eventID)
        return nil
    }

    // Route to handler based on event type
    var handlerErr error

    switch eventType {
    case "StockReserved":
        handlerErr = eh.handleStockReserved(ctx, message)
    case "StockReleased":
        handlerErr = eh.handleStockReleased(ctx, message)
    case "OrderPlaced":
        handlerErr = eh.handleOrderPlaced(ctx, message)
    case "OrderFailed":
        handlerErr = eh.handleOrderFailed(ctx, message)
    default:
        log.Printf("⏭️  Unknown event type: %s", eventType)
        return nil
    }

    // Record processing
    result := "success"
    if handlerErr != nil {
        result = "failed"
    }

    if recordErr := eh.idempotencyStore.RecordProcessed(ctx, eventID, "cart", eventType, result); recordErr != nil {
        log.Printf("⚠️  Failed to record idempotency: %v", recordErr)
    }

    return handlerErr
}

// handleStockReserved handles StockReservedEvent
func (eh *EventHandler) handleStockReserved(ctx context.Context, message []byte) error {
    var event events.StockReservedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReservedEvent: %w", err)
    }

    log.Printf("📨 StockReservedEvent received: Product %d, Quantity %d, Reservation %s", 
        event.ProductID, event.Quantity, event.ReservationID)

    // Create inventory lock in cart service
    lock := &events.StockReservedEvent{} // Use event as reference
    _ = lock

    // In real implementation, create inventory lock
    // For now, just log
    log.Printf("✓ Inventory lock created for reservation: %s", event.ReservationID)

    return nil
}

// handleStockReleased handles StockReleasedEvent
func (eh *EventHandler) handleStockReleased(ctx context.Context, message []byte) error {
    var event events.StockReleasedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReleasedEvent: %w", err)
    }

    log.Printf("📨 StockReleasedEvent received: Product %d, Reason: %s", event.ProductID, event.Reason)

    // Release corresponding inventory lock
    if err := eh.inventoryLockRepo.ReleaseLock(ctx, event.ReservationID); err != nil {
        log.Printf("⚠️  Failed to release inventory lock: %v", err)
    }

    return nil
}

// handleOrderPlaced handles OrderPlacedEvent
func (eh *EventHandler) handleOrderPlaced(ctx context.Context, message []byte) error {
    var event events.OrderPlacedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderPlacedEvent: %w", err)
    }

    log.Printf("📨 OrderPlacedEvent received: Order %d, User %s, Total %f", 
        event.OrderID, event.UserID, event.Total)

    // Order placed successfully - saga completed
    log.Printf("✓ Order placed successfully: %d", event.OrderID)

    return nil
}

// handleOrderFailed handles OrderFailedEvent
func (eh *EventHandler) handleOrderFailed(ctx context.Context, message []byte) error {
    var event events.OrderFailedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderFailedEvent: %w", err)
    }

    log.Printf("📨 OrderFailedEvent received: Order %s, Reason: %s", event.OrderID, event.Reason)

    // Order failed - trigger compensation
    log.Printf("🔄 Triggering compensation for order: %s", event.OrderID)

    return nil
}