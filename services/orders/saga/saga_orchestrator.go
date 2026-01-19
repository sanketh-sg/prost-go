package saga

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strconv"

    "github.com/google/uuid"
    "github.com/sanketh-sg/prost/services/orders/models"
    sharedmodels "github.com/sanketh-sg/prost/shared/models"
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
        log.Printf("Failed to check idempotency: %v", err)
    }

    if processed {
        log.Printf("Event %s already processed, skipping", eventID)
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
    case "OrderConfirmed":
        handlerErr = so.handleOrderConfirmed(ctx, message)
    case "OrderFailed":
        handlerErr = so.handleOrderFailed(ctx, message)
    case "OrderCancelled":
        handlerErr = so.handleOrderCancelled(ctx, message)
    default:
        log.Printf("Unknown event type: %s", eventType)
        return nil
    }

    // Record processing
    result := "success"
    if handlerErr != nil {
        result = "failed"
    }

    if recordErr := so.idempotencyStore.RecordProcessed(ctx, eventID, "orders", eventType, result); recordErr != nil {
        log.Printf("Failed to record idempotency: %v", recordErr)
    }

    return handlerErr
}

// handleCartCheckoutInitiated handles CartCheckoutInitiatedEvent (saga initiator)
func (so *SagaOrchestrator) handleCartCheckoutInitiated(ctx context.Context, message []byte) error {
    var event events.CartCheckoutInitiatedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal CartCheckoutInitiatedEvent: %w", err)
    }

    log.Printf("CartCheckoutInitiatedEvent received: Cart %s, User %s, Total %f", event.CartID, event.UserID, event.Total)

    // Get or create saga state
    correlationID := event.CorrelationID
    saga, err := so.sagaRepo.GetSagaState(ctx, correlationID)
    if err != nil {
        // Create new saga (if first time seeing this correlation ID)
        log.Printf("Creating new saga for correlation_id: %s", correlationID)
        saga = models.NewSagaState(event.CartID, event.UserID, correlationID)
        saga.Payload["items"] = event.Items
        saga.Payload["total"] = event.Total

        if err := so.sagaRepo.CreateSagaState(ctx, saga); err != nil {
            return fmt.Errorf("failed to create saga state: %w", err)
        }
    }

    // Step 1: Create order (pending state)
    // orderID := int64(uuid.New().ID()[:8])
	orderID := int64(uuid.New().ID())

    order := models.NewOrder(event.UserID, event.CartID, orderID, event.Total, correlationID)
    order.Status = "pending"

    if err := so.orderRepo.CreateOrder(ctx, order); err != nil {
        log.Printf("Failed to create order: %v", err)
        // Publish OrderFailedEvent to trigger compensation
        failedEvent := events.OrderFailedEvent{
            BaseEvent: events.NewBaseEvent("OrderFailed", strconv.FormatInt(orderID, 10), "order", correlationID),
            OrderID:   strconv.FormatInt(orderID, 10),
            Reason:    "failed to create order record",
        }
        if pubErr := so.eventPublisher.PublishOrderEvent(ctx, failedEvent); pubErr != nil {
            log.Printf("Failed to publish OrderFailedEvent: %v", pubErr)
        }
        return err
    }

    log.Printf("Order created: %d", orderID)

    // Update saga with order ID
    if err := so.sagaRepo.UpdateSagaOrderID(ctx, correlationID, orderID); err != nil {
        log.Printf("Failed to update saga with order_id: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    // Update saga status to order_created
    if err := so.sagaRepo.UpdateSagaStatus(ctx, correlationID, "order_created"); err != nil {
        log.Printf("Failed to update saga status: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    // Step 2: Publish OrderPlacedEvent (triggers inventory reservation in products service)
    orderCreatedEvent := events.OrderCreatedEvent{
        BaseEvent: events.NewBaseEvent("OrderCreated", strconv.FormatInt(orderID, 10), "order", correlationID),
        OrderID:   orderID,
        UserID:    event.UserID,
        Total:     event.Total,
        Items:     event.Items,
    }

    if err := so.eventPublisher.PublishOrderEvent(ctx, orderCreatedEvent); err != nil {
        log.Printf("Failed to publish OrderCreatedEvent: %v", err)
        return err
    }

    log.Printf("OrderCreatedEvent published for order: %d", orderID)
    // Update saga to waiting for inventory
    if err := so.sagaRepo.UpdateSagaStatus(ctx, correlationID, "checking_inventory"); err != nil {
        log.Printf("Failed to update saga status: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    return nil
}

// handleStockReserved handles StockReservedEvent (saga step 2)
func (so *SagaOrchestrator) handleStockReserved(ctx context.Context, message []byte) error {
    var event events.StockReservedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReservedEvent: %w", err)
    }

    log.Printf("StockReservedEvent received: Product %d, Quantity %d, Reservation %s", 
        event.ProductID, event.Quantity, event.ReservationID)

    // Get saga to check inventory status
    saga, err := so.sagaRepo.GetSagaState(ctx, event.CorrelationID)
    if err != nil || saga == nil {
        return fmt.Errorf("saga not found: %s", event.CorrelationID)
    }


    // Create inventory reservation in orders schema
    if event.OrderID > 0 {
        res := models.NewInventoryReservation(event.OrderID, event.ProductID, event.Quantity, event.ReservationID)
        if err := so.inventoryResRepo.CreateReservation(ctx, res); err != nil {
            log.Printf("Failed to create inventory reservation: %v", err)
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
            log.Printf("Failed to create compensation log: %v", err)
        }
    }

    // Get order to transition to placed
    orderID, ok := saga.Payload["order_id"].(int64)
    if !ok {
        return fmt.Errorf("order_id not found in saga")
    }
    // Update it to order placed
    if err := so.orderRepo.UpdateOrderStatus(ctx, orderID, "placed"); err != nil {
        log.Printf("Failed to update order status to placed: %v", err)
        return err
    }

    log.Printf("Order transitioned to PLACED: %d (all inventory reserved)", orderID)


    // Step 3: Publish OrderPlacedEvent (now order is officially placed with confirmed inventory)
    orderPlacedEvent := events.OrderPlacedEvent{
        BaseEvent: events.NewBaseEvent("OrderPlaced", strconv.FormatInt(orderID, 10), "order", event.CorrelationID),
        OrderID:   orderID,
        UserID:    saga.Payload["user_id"].(string),
        Total:     saga.Payload["total"].(float64),
        Items:     saga.Payload["items"].([]sharedmodels.OrderItem),
    }

    if err := so.eventPublisher.PublishOrderEvent(ctx, orderPlacedEvent); err != nil {
        log.Printf("Failed to publish OrderPlacedEvent: %v", err)
    }

    log.Printf("✓ OrderPlacedEvent published: %d", orderID)

    // Update saga status
    if err := so.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "order_placed"); err != nil {
        log.Printf("Failed to update saga status: %v", err)
    }

    return nil
}


// handleStockReleased handles StockReleasedEvent (saga compensation)
func (so *SagaOrchestrator) handleStockReleased(ctx context.Context, message []byte) error {
    var event events.StockReleasedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal StockReleasedEvent: %w", err)
    }

    log.Printf("StockReleasedEvent received: Product %d, Reason: %s", event.ProductID, event.Reason)

    // Release inventory reservation
    if err := so.inventoryResRepo.ReleaseReservation(ctx, event.ReservationID); err != nil {
        log.Printf("Failed to release inventory reservation: %v", err)
    }

    return nil
}


// handleOrderConfirmed handles OrderConfirmedEvent (saga step 3 - confirmation)
// Why: When all items are confirmed and payment succeeds, mark saga as completed
func (so *SagaOrchestrator) handleOrderConfirmed(ctx context.Context, message []byte) error {
    var event events.OrderConfirmedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderConfirmedEvent: %w", err)
    }

    log.Printf("Processing OrderConfirmedEvent: OrderID=%d, CorrelationID=%s", event.OrderID, event.CorrelationID)

    // Update order status to "confirmed"
    if err := so.orderRepo.UpdateOrderStatus(ctx, event.OrderID, "confirmed"); err != nil {
        log.Printf("Failed to update order status to confirmed: %v", err)
        return fmt.Errorf("failed to update order status: %w", err)
    }

    log.Printf("Order status updated to confirmed: %d", event.OrderID)

    // Update saga status to "completed"
    if err := so.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "completed"); err != nil {
        log.Printf("Failed to update saga status to completed: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    log.Printf("✓ Saga completed for order: %d", event.OrderID)

    return nil
}

// handleOrderFailed handles OrderFailedEvent (saga failure/compensation)
// Why: When order fails at any step, release reserved inventory and mark order as failed
func (so *SagaOrchestrator) handleOrderFailed(ctx context.Context, message []byte) error {
    var event events.OrderFailedEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderFailedEvent: %w", err)
    }

    orderID, err := strconv.ParseInt(event.OrderID, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid order ID: %w", err)
    }

    log.Printf("Processing OrderFailedEvent: OrderID=%s, Reason=%s, CorrelationID=%s", 
        event.OrderID, event.Reason, event.CorrelationID)

    // Update order status to "failed"
    if err := so.orderRepo.UpdateOrderStatus(ctx, orderID, "failed"); err != nil {
        log.Printf("Failed to update order status to failed: %v", err)
        return fmt.Errorf("failed to update order status: %w", err)
    }

    log.Printf("✓ Order status updated to failed: %d", orderID)

    // Get all compensation logs for this order
    compensationLogs, err := so.compensationRepo.GetCompensationLogsByOrderID(ctx, orderID)
    if err != nil {
        log.Printf("Failed to get compensation logs: %v", err)
    }

    // Execute compensation in reverse order (LIFO)
    for i := len(compensationLogs) - 1; i >= 0; i-- {
        compLog := compensationLogs[i]
        log.Printf("Executing compensation: %s for order %d", compLog.CompensationEvent, orderID)

        // The compensation is already tracked; Products service handles actual stock release
        // via StockReleasedEvent from order failure
    }

    // Update saga status to "failed"
    if err := so.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "failed"); err != nil {
        log.Printf("Failed to update saga status to failed: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    log.Printf("✓ Saga marked as failed for order: %d, Reason: %s", orderID, event.Reason)

    return nil
}

// handleOrderCancelled handles OrderCancelledEvent (saga cancellation)
// Why: When user/admin cancels order, release reserved inventory and mark order as cancelled
func (so *SagaOrchestrator) handleOrderCancelled(ctx context.Context, message []byte) error {
    var event events.OrderCancelledEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return fmt.Errorf("failed to unmarshal OrderCancelledEvent: %w", err)
    }

    orderID, err := strconv.ParseInt(event.OrderID, 10, 64)
    if err != nil {
        return fmt.Errorf("invalid order ID: %w", err)
    }

    log.Printf("Processing OrderCancelledEvent: OrderID=%s, Reason=%s, CorrelationID=%s", 
        event.OrderID, event.Reason, event.CorrelationID)

    // Update order status to "cancelled"
    if err := so.orderRepo.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
        log.Printf("Failed to update order status to cancelled: %v", err)
        return fmt.Errorf("failed to update order status: %w", err)
    }

    log.Printf("Order status updated to cancelled: %d", orderID)

    // Get all compensation logs for this order
    compensationLogs, err := so.compensationRepo.GetCompensationLogsByOrderID(ctx, orderID)
    if err != nil {
        log.Printf("Failed to get compensation logs: %v", err)
    }

    // Execute compensation in reverse order (LIFO)
    for i := len(compensationLogs) - 1; i >= 0; i-- {
        compLog := compensationLogs[i]
        log.Printf("Executing compensation: %s for order %d", compLog.CompensationEvent, orderID)

        // The compensation is already tracked; Products service handles actual stock release
        // via StockReleasedEvent from order cancellation
    }

    // Update saga status to "cancelled"
    if err := so.sagaRepo.UpdateSagaStatus(ctx, event.CorrelationID, "cancelled"); err != nil {
        log.Printf("Failed to update saga status to cancelled: %v", err)
        return fmt.Errorf("failed to update saga status: %w", err)
    }

    log.Printf("aga marked as cancelled for order: %d, Reason: %s", orderID, event.Reason)

    return nil
}