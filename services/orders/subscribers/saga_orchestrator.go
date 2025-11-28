package subscribers

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strconv"

    "github.com/google/uuid"
    "github.com/sanketh-sg/prost/services/orders/models"
    "github.com/sanketh-sg/prost/services/orders/repository"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/events"
    "github.com/sanketh-sg/prost/shared/messaging"
)

// SagaOrchestrator orchestrates order creation saga
type SagaOrchestrator struct {
    orderRepo         *repository.OrderRepository
    sagaRepo          *repository.SagaStateRepository
    compensationRepo  *repository.CompensationLogRepository
    inventoryResRepo  *repository.InventoryReservationRepository
    idempotencyStore  *db.IdempotencyStore
    eventPublisher    *messaging.Publisher
}

// NewSagaOrchestrator creates new saga orchestrator
func NewSagaOrchestrator(
    orderRepo *repository.OrderRepository,
    sagaRepo *repository.SagaStateRepository,
    compensationRepo *repository.CompensationLogRepository,
    inventoryResRepo *repository.InventoryReservationRepository,
    idempotencyStore *db.IdempotencyStore,
    eventPublisher *messaging.Publisher,
) *SagaOrchestrator {
    return &SagaOrchestrator{
        orderRepo:        orderRepo,
        sagaRepo:         sagaRepo,
        compensationRepo: compensationRepo,
        inventoryResRepo: inventoryResRepo,
        idempotencyStore: idempotencyStore,
        eventPublisher:   eventPublisher,
    }
}

// HandleEvent processes incoming events for saga
func (so *SagaOrchestrator) HandleEvent(ctx context.Context, message []byte) error {
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
    processed, err := so.idempotencyStore.IsProcessed(ctx, eventID, "orders")
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
    case "CartCheckoutInitiated":
        handlerErr = so.handleCartCheckoutInitiated(ctx, message)
    case "StockReserved":
        handlerErr = so.handleStockReserved(ctx, message)
    case "StockReleased":
        handlerErr = so.handleStockReleased(ctx, message)
    default:
        log.Printf("⏭️  Unknown event type: %s", eventType)
        return nil
    }

    // Record processing
    result := "success"
    if handlerErr != nil {
        result = "failed"
    }

    if recordErr := so.idempotencyStore.RecordProcessed(ctx, eventID, "orders", eventType, result); recordErr != nil {
        log.Printf("⚠️  Failed to record idempotency: %v", recordErr)
    }

    return handlerErr
}

// handleCartCheckoutInitiated handles CartCheckoutInitiatedEvent (saga initiator)
func (so *SagaOrchestrator) handleCartCheckoutInitiated(ctx context.Context, message []byte) error {
    var event events.CartCheckoutInitiatedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal CartCheckoutInitiatedEvent: %w", err)
    }

    log.Printf("📨 CartCheckoutInitiatedEvent received: Cart %s, User %s, Total %f", event.CartID, event.UserID, event.Total)

    // Get or create saga state
    correlationID := event.CorrelationID
    saga, err := so.sagaRepo.GetSagaState(ctx, correlationID)
    if err != nil {
        // Create new saga (if first time seeing this correlation ID)
        log.Printf("Creating new saga for correlation_id: %s", correlationID)
        saga = models.NewSagaState(event.CartID, event.UserID, correlationID)
        saga.Payload["items"] = []map[string]interface{}{} // Will be populated later
        saga.Payload["total"] = event.Total

        if err := so.sagaRepo.CreateSagaState(ctx, saga); err != nil {
            return fmt.Errorf("failed to create saga state: %w", err)
        }
    }

    // Step 1: Create order (pending state)
    // orderID := int64(uuid.New().ID()[:8])
	orderID := int64(uuid.New().ID())

    order := models.NewOrder(event.UserID, event.CartID, orderID, event.Total, correlationID)

    if err := so.orderRepo.CreateOrder(ctx, order); err != nil {
        log.Printf("❌ Failed to create order: %v", err)
        // Publish OrderFailedEvent to trigger compensation
        failedEvent := events.OrderFailedEvent{
            BaseEvent: events.NewBaseEvent("OrderFailed", strconv.FormatInt(orderID, 10), "order", correlationID),
            OrderID:   strconv.FormatInt(orderID, 10),
            Reason:    "failed to create order record",
        }
        if pubErr := so.eventPublisher.PublishOrderEvent(ctx, failedEvent); pubErr != nil {
            log.Printf("⚠️  Failed to publish OrderFailedEvent: %v", pubErr)
        }
        return err
    }

    log.Printf("✓ Order created: %d", orderID)

    // Update saga with order ID
    if err := so.sagaRepo.UpdateSagaOrderID(ctx, correlationID, orderID); err != nil {
        log.Printf("⚠️  Failed to update saga with order_id: %v", err)
    }

    // Update saga status to CART_VALIDATED
    if err := so.sagaRepo.UpdateSagaStatus(ctx, correlationID, "cart_validated"); err != nil {
        log.Printf("⚠️  Failed to update saga status: %v", err)
    }

    // Step 2: Publish OrderPlacedEvent (triggers inventory reservation in products service)
    orderPlacedEvent := events.OrderPlacedEvent{
        BaseEvent: events.NewBaseEvent("OrderPlaced", strconv.FormatInt(orderID, 10), "order", correlationID),
        OrderID:   orderID,
        UserID:    event.UserID,
        Total:     event.Total,
        // Items:     []struct {
        //     ProductID int64
        //     Quantity  int
        //     Price     float64
        // }{}, // TODO: Get from cart context
		Items:     []struct {
            ProductID int64   `json:"product_id"`
            Quantity  int     `json:"quantity"`
            Price     float64 `json:"price"`
        }{}, // TODO: Get from cart context
    }

    if err := so.eventPublisher.PublishOrderEvent(ctx, orderPlacedEvent); err != nil {
        log.Printf("❌ Failed to publish OrderPlacedEvent: %v", err)
        return err
    }

    log.Printf("✓ OrderPlacedEvent published for order: %d", orderID)

    return nil
}

// handleStockReserved handles StockReservedEvent (saga step 2)
func (so *SagaOrchestrator) handleStockReserved(ctx context.Context, message []byte) error {
    var event events.StockReservedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReservedEvent: %w", err)
    }

    log.Printf("📨 StockReservedEvent received: Product %d, Quantity %d, Reservation %s", 
        event.ProductID, event.Quantity, event.ReservationID)

    // Create inventory reservation in orders schema
    if event.OrderID > 0 {
        res := models.NewInventoryReservation(event.OrderID, event.ProductID, event.Quantity, event.ReservationID)
        if err := so.inventoryResRepo.CreateReservation(ctx, res); err != nil {
            log.Printf("⚠️  Failed to create inventory reservation: %v", err)
        }

        // Add to compensation log (in case we need to rollback)
        compensation := models.NewCompensationLog(
            event.OrderID,
            event.CorrelationID,
            "StockReleased",
            map[string]interface{}{
                "reservation_id": event.ReservationID,
                "product_id":     event.ProductID,
                "quantity":       event.Quantity,
            },
        )
        if err := so.compensationRepo.CreateCompensationLog(ctx, compensation); err != nil {
            log.Printf("⚠️  Failed to create compensation log: %v", err)
        }

        // Update saga status
        if err := so.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "inventory_reserved"); err != nil {
            log.Printf("⚠️  Failed to update saga status: %v", err)
        }
    }

    return nil
}

// handleStockReleased handles StockReleasedEvent (saga compensation)
func (so *SagaOrchestrator) handleStockReleased(ctx context.Context, message []byte) error {
    var event events.StockReleasedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReleasedEvent: %w", err)
    }

    log.Printf("📨 StockReleasedEvent received: Product %d, Reason: %s", event.ProductID, event.Reason)

    // Release inventory reservation
    if err := so.inventoryResRepo.ReleaseReservation(ctx, event.ReservationID); err != nil {
        log.Printf("⚠️  Failed to release inventory reservation: %v", err)
    }

    return nil
}