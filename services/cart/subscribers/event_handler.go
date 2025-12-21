package subscribers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sanketh-sg/prost/services/cart/models"
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
// Why: Events from Products and Orders services need to update cart state
// Events can be: StockReserved, StockReleased, OrderPlaced, OrderFailed
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

    // Check idempotency - prevent processing same event twice
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

// handleStockReserved handles StockReservedEvent from Products service
// Why: When Products service reserves inventory, we create an inventory lock in cart
// This prevents double-selling if multiple orders try to buy the same item
func (eh *EventHandler) handleStockReserved(ctx context.Context, message []byte) error {
    var event events.StockReservedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReservedEvent: %w", err)
    }

    log.Printf("📨 StockReservedEvent received: Product %d, Quantity %d, Reservation %s",
        event.ProductID, event.Quantity, event.ReservationID)

    // If event has order_id, create inventory lock in our database
    if event.OrderID > 0 {
        lock := &models.InventoryLock{
            CartID:        fmt.Sprintf("order-%d", event.OrderID), // Link to order
            ProductID:     event.ProductID,
            Quantity:      event.Quantity,
            ReservationID: event.ReservationID,
            Status:        "locked", // locked = inventory reserved in products service
            LockedAt:      time.Now(),
        }

        if err := eh.inventoryLockRepo.CreateLock(ctx, lock); err != nil {
            log.Printf("❌ Failed to create inventory lock: %v", err)
            return fmt.Errorf("failed to create inventory lock: %w", err)
        }

        log.Printf("✓ Inventory lock created: Product %d, Reservation %s", event.ProductID, event.ReservationID)

        // Update saga state to reflect inventory locked
        if err := eh.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "inventory_locked"); err != nil {
            log.Printf("⚠️  Failed to update saga status: %v", err)
        }
    }

    return nil
}

// handleStockReleased handles StockReleasedEvent (compensation)
// Why: When an order fails or is cancelled, Products service releases inventory
// We need to remove the lock from our records
func (eh *EventHandler) handleStockReleased(ctx context.Context, message []byte) error {
    var event events.StockReleasedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReleasedEvent: %w", err)
    }

    log.Printf("📨 StockReleasedEvent received: Product %d, Reservation %s, Reason: %s",
        event.ProductID, event.ReservationID, event.Reason)

    // Release (remove) the inventory lock
    if err := eh.inventoryLockRepo.ReleaseLock(ctx, event.ReservationID); err != nil {
        log.Printf("❌ Failed to release inventory lock: %v", err)
        return fmt.Errorf("failed to release inventory lock: %w", err)
    }

    log.Printf("✓ Inventory lock released: Reservation %s (Reason: %s)", event.ReservationID, event.Reason)

    // If this is due to order failure, update saga status
    if event.Reason == "order_failed" || event.Reason == "order_cancelled" {
        if err := eh.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "failed"); err != nil {
            log.Printf("⚠️  Failed to update saga status to failed: %v", err)
        }
    }

    return nil
}

// handleOrderPlaced handles OrderPlacedEvent from Orders service
// Why: Order was successfully created and inventory is reserved
// We can now mark the saga as complete and clear the cart
func (eh *EventHandler) handleOrderPlaced(ctx context.Context, message []byte) error {
    var event events.OrderPlacedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderPlacedEvent: %w", err)
    }

    log.Printf("📨 OrderPlacedEvent received: Order %d, User %s, Total %f",
        event.OrderID, event.UserID, event.Total)

    // Update saga state to confirmed
    if err := eh.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "order_confirmed"); err != nil {
        log.Printf("❌ Failed to update saga status: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    log.Printf("✓ Saga marked as confirmed: %s", event.CorrelationID)

    return nil
}

// handleOrderFailed handles OrderFailedEvent (compensation)
// Why: Order creation failed for some reason (payment, inventory issue, etc.)
// We need to release all inventory locks and mark saga as compensating
func (eh *EventHandler) handleOrderFailed(ctx context.Context, message []byte) error {
    var event events.OrderFailedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderFailedEvent: %w", err)
    }

    log.Printf("📨 OrderFailedEvent received: Order %s, Reason: %s", event.OrderID, event.Reason)

    // Get the saga to find correlation ID
    orderID, err := strconv.ParseInt(event.OrderID, 10, 64)
    if err == nil {
        // Find saga by correlation ID from event
        sagaStatus := "compensation_in_progress"

        if err := eh.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, sagaStatus); err != nil {
            log.Printf("❌ Failed to update saga status to compensating: %v", err)
            return fmt.Errorf("failed to update saga status: %w", err)
        }

        log.Printf("✓ Saga marked for compensation: %s (Order %d, Reason: %s)", 
            event.CorrelationID, orderID, event.Reason)

        // Note: Products service will handle releasing inventory via RabbitMQ
        // We just mark the saga state for our records
    }

    return nil
}