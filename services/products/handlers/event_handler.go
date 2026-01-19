package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sanketh-sg/prost/services/products/models"
	"github.com/sanketh-sg/prost/services/products/repository"
	"github.com/sanketh-sg/prost/shared/db"
	"github.com/sanketh-sg/prost/shared/events"
	"github.com/sanketh-sg/prost/shared/messaging"
)

// EventHandler handles incoming events for products service
type EventHandler struct {
	inventoryRepo    *repository.InventoryReservationRepository
	idempotencyStore *db.IdempotencyStore
    eventPublisher   *messaging.Publisher
}

// NewEventHandler creates new event handler
func NewEventHandler(
	inventoryRepo *repository.InventoryReservationRepository,
	idempotencyStore *db.IdempotencyStore,
    eventPublisher   *messaging.Publisher,
) *EventHandler {
	return &EventHandler{
		inventoryRepo:    inventoryRepo,
		idempotencyStore: idempotencyStore,
        eventPublisher: eventPublisher,
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

	// Check idempotency - prevent processing same event twice
	processed, err := eh.idempotencyStore.IsProcessed(ctx, eventID, "products")
	if err != nil {
		log.Printf("Failed to check idempotency: %v", err)
	}

	if processed {
		log.Printf("Event %s already processed, skipping", eventID)
		return nil
	}

    // Route to handler based on event type
    var handlerErr error

    switch eventType {
    case "OrderCreated":
        handlerErr = eh.handleOrderCreated(ctx, message)
    case "OrderConfirmed":
        handlerErr = eh.handleOrderConfirmed(ctx, message)
    case "OrderFailed":
        handlerErr = eh.handleOrderFailed(ctx, message)
    case "OrderCancelled":
        handlerErr = eh.handleOrderCancelled(ctx, message)
    default:
        log.Printf("Unknown event type: %s, skipping", eventType)
        return nil
    }

	// Record processing
	result := "success"
	if handlerErr != nil {
		result = "failed"
	}

	if recordErr := eh.idempotencyStore.RecordProcessed(ctx, eventID, "products", eventType, result); recordErr != nil {
		log.Printf("Failed to record idempotency: %v", recordErr)
	}

	return handlerErr
}

// handleOrderPlaced handles OrderPlacedEvent
// Why: When order is placed, we need to reserve the stock
// This prevents double-selling when multiple orders try to buy same item
func (eh *EventHandler) handleOrderCreated(ctx context.Context, message []byte) error {
    var event events.OrderCreatedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderCreatedEvent: %w", err)
    }

    log.Printf("Processing OrderCreatedEvent: OrderID=%d, Items=%d", event.OrderID, len(event.Items))

    insufficientInventory := false
    // First: Check if all items have sufficient inventory
    for _, item := range event.Items {
        inventory, err := eh.inventoryRepo.GetProductInventory(ctx, item.ProductID)
        if err != nil || inventory == nil || inventory.AvailableQuantity < item.Quantity {
            log.Printf("Insufficient inventory for product %d: need %d, have %d", 
                item.ProductID, item.Quantity, 
                func() int { //anonymous function to get available quantity
                    if inventory != nil {
                        return inventory.AvailableQuantity
                    }
                    return 0
                }())
                insufficientInventory = true
            break
        }
    }

    if insufficientInventory{
        // Publish OrderFailedEvent to trigger compensation
            failedEvent := events.OrderFailedEvent{
                BaseEvent: events.NewBaseEvent("OrderFailed", fmt.Sprintf("%d", event.OrderID), "order", event.CorrelationID),
                OrderID:   fmt.Sprintf("%d", event.OrderID),
                Reason:    "Insufficient inventory for product",
            }
            if err := eh.eventPublisher.PublishProductEvent(ctx, failedEvent); err != nil {
                log.Printf("Failed to publish OrderFailedEvent: %v", err)
            }
            return fmt.Errorf("insufficient inventory for products")
    } 
    // Reserve stock for each item in the order
    for _, item := range event.Items {
        reservation := &models.InventoryReservation{
            ProductID:     item.ProductID,
            Quantity:      item.Quantity,
            OrderID:       event.OrderID,
            ReservationID: fmt.Sprintf("res-%d-%d", event.OrderID, item.ProductID), // Generate unique ID
            Status:        "reserved",
            CreatedAt: time.Now(),
            ExpiresAt: time.Now().Add(5*time.Minute),
        }

        if err := eh.inventoryRepo.CreateReservation(ctx, reservation); err != nil {
            // Cleanup: Release already-reserved items
            eh.releaseReservationsForOrder(ctx, event.OrderID)
            
            // Publish ONE OrderFailedEvent
            failedEvent := events.OrderFailedEvent{
                BaseEvent:    events.NewBaseEvent("OrderFailed", fmt.Sprintf("%d", event.OrderID), "order", event.CorrelationID),
                OrderID:      fmt.Sprintf("%d", event.OrderID),
                Reason:       fmt.Sprintf("failed to reserve inventory for product %d", item.ProductID),
            }
            if err := eh.eventPublisher.PublishProductEvent(ctx, failedEvent); err != nil {
                log.Printf("Failed to publish OrderFailedEvent: %v", err)
            }
            return fmt.Errorf("failed to create reservation for product %d: %w", item.ProductID, err)
        }

        log.Printf("Reserved %d units of product %d for order %d", item.Quantity, item.ProductID, event.OrderID)

        // Publish StockReservedEvent for each item
        stockEvent := events.StockReservedEvent{
            BaseEvent:     events.NewBaseEvent("StockReserved", fmt.Sprintf("%d", item.ProductID), "product", fmt.Sprintf("%d", event.OrderID)),
            ProductID:     item.ProductID,
            Quantity:      item.Quantity,
            OrderID:       event.OrderID,
            ReservationID: reservation.ReservationID,
        }

        if err := eh.eventPublisher.PublishProductEvent(ctx, stockEvent); err != nil {
            log.Printf("Failed to publish StockReservedEvent: %v", err)
            // Don't fail - idempotency will handle retry
        }
    }

    return nil
}


// handleOrderConfirmed handles OrderConfirmedEvent
// Why: When order is confirmed, mark the reservation as "confirmed"/"sold"
// This indicates the stock has been permanently allocated to the order
func (eh *EventHandler) handleOrderConfirmed(ctx context.Context, message []byte) error {
    var event events.OrderConfirmedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderConfirmedEvent: %w", err)
    }

    log.Printf("✓ Processing OrderConfirmedEvent: OrderID=%d", event.OrderID)

    // Update reservation status to "confirmed"
    if err := eh.inventoryRepo.UpdateReservationStatusByOrderID(ctx, fmt.Sprintf("%d", event.OrderID), "confirmed"); err != nil {
        log.Printf("Failed to update reservation status to confirmed: %v", err)
        return fmt.Errorf("failed to update reservation status: %w", err)
    }

    log.Printf("✓ Reservation confirmed for order: %d", event.OrderID)
    return nil
}

// handleOrderFailed handles OrderFailedEvent
// Why: When order fails, release the reserved inventory
// This allows stock to be sold to other customers
func (eh *EventHandler) handleOrderFailed(ctx context.Context, message []byte) error {
    var event events.OrderFailedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderFailedEvent: %w", err)
    }

    log.Printf("Processing OrderFailedEvent: OrderID=%s, Reason=%s", event.OrderID, event.Reason)

    // Get all reservations for this order
    orderID, err := strconv.ParseInt(event.OrderID, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid order ID: %w", err)
    }

    reservations, err := eh.inventoryRepo.GetReservationsByOrderID(ctx, orderID)
    if err != nil {
        log.Printf("Failed to get reservations for order: %v", err)
        return fmt.Errorf("failed to get reservations: %w", err)
    }

    // Release each reservation
    for _, res := range reservations {
        if err := eh.inventoryRepo.ReleaseReservation(ctx, res.ReservationID); err != nil {
            log.Printf(" Failed to release reservation %s: %v", res.ReservationID, err)
            return fmt.Errorf("failed to release reservation: %w", err)
        }

        // Publish StockReleasedEvent
        stockEvent := events.StockReleasedEvent{
            BaseEvent:     events.NewBaseEvent("StockReleased", fmt.Sprintf("%d", res.ProductID), "product", event.OrderID),
            ProductID:     res.ProductID,
            Quantity:      res.Quantity,
            ReservationID: res.ReservationID,
            Reason:        event.Reason,
        }

        if err := eh.eventPublisher.PublishProductEvent(ctx, stockEvent); err != nil {
            log.Printf("Failed to publish StockReleasedEvent: %v", err)
        }

        log.Printf("Released %d units of product %d for failed order %s", res.Quantity, res.ProductID, event.OrderID)
    }

    return nil
}

// handleOrderCancelled handles OrderCancelledEvent
// Why: When order is cancelled by user/admin, release the reserved inventory
// This allows the stock to be allocated to other orders
func (eh *EventHandler) handleOrderCancelled(ctx context.Context, message []byte) error {
    var event events.OrderCancelledEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderCancelledEvent: %w", err)
    }

    log.Printf("Processing OrderCancelledEvent: OrderID=%s, Reason=%s", event.OrderID, event.Reason)

    // Get all reservations for this order
    orderID, err := strconv.ParseInt(event.OrderID, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid order ID: %w", err)
    }

    reservations, err := eh.inventoryRepo.GetReservationsByOrderID(ctx, orderID)
    if err != nil {
        log.Printf("Failed to get reservations for order: %v", err)
        return fmt.Errorf("failed to get reservations: %w", err)
    }

    // Release each reservation
    for _, res := range reservations {
        if err := eh.inventoryRepo.ReleaseReservation(ctx, res.ReservationID); err != nil {
            log.Printf("Failed to release reservation %s: %v", res.ReservationID, err)
            return fmt.Errorf("failed to release reservation: %w", err)
        }

        // Publish StockReleasedEvent
        stockEvent := events.StockReleasedEvent{
            BaseEvent:     events.NewBaseEvent("StockReleased", fmt.Sprintf("%d", res.ProductID), "product", event.OrderID),
            ProductID:     res.ProductID,
            Quantity:      res.Quantity,
            ReservationID: res.ReservationID,
            Reason:        "order_cancelled", // Clear reason
        }

        if err := eh.eventPublisher.PublishProductEvent(ctx, stockEvent); err != nil {
            log.Printf("Failed to publish StockReleasedEvent: %v", err)
        }

        log.Printf("Released %d units of product %d for cancelled order %s", res.Quantity, res.ProductID, event.OrderID)
    }

    return nil
}

// releaseReservationsForOrder releases all reservations for an order
// Used when order fails after partial reservations
func (eh *EventHandler) releaseReservationsForOrder(ctx context.Context, orderID int64) {
    reservations, err := eh.inventoryRepo.GetReservationsByOrderID(ctx, orderID)
    if err != nil {
        log.Printf("Failed to get reservations for cleanup: %v", err)
        return
    }

    for _, res := range reservations {
        if err := eh.inventoryRepo.ReleaseReservation(ctx, res.ReservationID); err != nil {
            log.Printf("Failed to release reservation %s during cleanup: %v", res.ReservationID, err)
        }
    }
}