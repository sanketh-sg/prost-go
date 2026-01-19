
package handlers

import (
    "context"
    "log"
    "net/http"
    "strconv"
    "time"
    "fmt"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/orders/models"
    "github.com/sanketh-sg/prost/services/orders/repository"
    "github.com/sanketh-sg/prost/services/orders/saga"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/messaging"
    "github.com/sanketh-sg/prost/shared/events"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
    orderRepo         *repository.OrderRepository
    sagaRepo          *repository.SagaStateRepository
    compensationRepo  *repository.CompensationLogRepository
    inventoryResRepo  *repository.InventoryReservationRepository
    idempotencyStore  *db.IdempotencyStore
    eventPublisher    *messaging.Publisher
    sagaOrchestrator  *saga.SagaOrchestrator
}

// NewOrderHandler creates new order handler
func NewOrderHandler(
    orderRepo *repository.OrderRepository,
    sagaRepo *repository.SagaStateRepository,
    compensationRepo *repository.CompensationLogRepository,
    inventoryResRepo *repository.InventoryReservationRepository,
    idempotencyStore *db.IdempotencyStore,
    eventPublisher *messaging.Publisher,
    sagaOrchestrator *saga.SagaOrchestrator,
) *OrderHandler {
    return &OrderHandler{
        orderRepo:        orderRepo,
        sagaRepo:         sagaRepo,
        compensationRepo: compensationRepo,
        inventoryResRepo: inventoryResRepo,
        idempotencyStore: idempotencyStore,
        eventPublisher:   eventPublisher,
        sagaOrchestrator: sagaOrchestrator,
    }
}

// Health handles health check
func (oh *OrderHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "orders",
        "time":    time.Now().UTC(),
    })
}

// GetOrder retrieves an order
func (oh *OrderHandler) GetOrder(c *gin.Context) {
    // ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    orderIDStr := c.Param("id")
    orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid order id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    order, err := oh.orderRepo.GetOrder(ctx, orderID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "order not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, order)
}

// GetOrders retrieves orders for a user
func (oh *OrderHandler) GetOrders(c *gin.Context) {
    // ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID := c.Query("user_id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "user_id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    orders, err := oh.orderRepo.GetOrdersByUserID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to get orders",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "orders": orders,
        "count":  len(orders),
    })
}

// GetSagaState retrieves saga state
func (oh *OrderHandler) GetSagaState(c *gin.Context) {
    // ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    correlationID := c.Param("correlation_id")
    if correlationID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "correlation_id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    saga, err := oh.sagaRepo.GetSagaState(ctx, correlationID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "saga not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, saga)
}

// CancelOrder cancels an order
func (oh *OrderHandler) CancelOrder(c *gin.Context) {
    // ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    orderIDStr := c.Param("id")
    orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid order id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    var req models.CancelOrderRequest
    //check if it is a valid cancel request
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get order
    order, err := oh.orderRepo.GetOrder(ctx, orderID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "order not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Cancel order
    if err := oh.orderRepo.CancelOrder(ctx, orderID); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to cancel order",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Get inventory reservations and release them
    reservations, err := oh.inventoryResRepo.GetReservationsByOrderID(ctx, orderID)
    if err == nil {
        for _, res := range reservations {
            if err := oh.inventoryResRepo.ReleaseReservation(ctx, res.ReservationID); err != nil {
                log.Printf("⚠️  Failed to release reservation: %v", err)
            }
        }
    }

    // Publish OrderCancelledEvent to trigger compensation (inventory release)
    cancelledEvent := events.OrderCancelledEvent{
        BaseEvent: events.NewBaseEvent("OrderCancelled", fmt.Sprintf("%d", orderID), "order", order.SagaCorrelationID),
        OrderID:   fmt.Sprintf("%d", orderID),
        Reason:    req.Reason, // provided by user
    }
    if err := oh.eventPublisher.PublishOrderEvent(ctx, cancelledEvent); err != nil {
        log.Printf("Failed to publish OrderCancelledEvent: %v", err)
    }

    log.Printf("✓ Order cancelled: %d, Reason: %s", orderID, req.Reason)

    c.JSON(http.StatusOK, gin.H{
        "message": "Order cancelled successfully",
        "order_id": orderID,
        "saga_correlation_id": order.SagaCorrelationID,
    })
}