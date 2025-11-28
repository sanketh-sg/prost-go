package handlers

import (
    "context"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sanketh-sg/prost/services/cart/models"
    "github.com/sanketh-sg/prost/services/cart/repository"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/events"
    "github.com/sanketh-sg/prost/shared/messaging"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
    cartRepo           *repository.CartRepository
    sagaRepo           *repository.SagaStateRepository
    inventoryLockRepo  *repository.InventoryLockRepository
    idempotencyStore   *db.IdempotencyStore
    eventPublisher     *messaging.Publisher
}

// NewCartHandler creates new cart handler
func NewCartHandler(
    cartRepo *repository.CartRepository,
    sagaRepo *repository.SagaStateRepository,
    inventoryLockRepo *repository.InventoryLockRepository,
    idempotencyStore *db.IdempotencyStore,
    eventPublisher *messaging.Publisher,
) *CartHandler {
    return &CartHandler{
        cartRepo:          cartRepo,
        sagaRepo:          sagaRepo,
        inventoryLockRepo: inventoryLockRepo,
        idempotencyStore:  idempotencyStore,
        eventPublisher:    eventPublisher,
    }
}

// Health handles health check
func (ch *CartHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "cart",
        "time":    time.Now().UTC(),
    })
}

// CreateCart creates a new cart
func (ch *CartHandler) CreateCart(c *gin.Context) {
    ctx := context.Background()

    var req models.CreateCartRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    cart := models.NewCart(req.UserID)
    if err := ch.cartRepo.CreateCart(ctx, cart); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create cart",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ Cart created: %s for user %s", cart.ID, cart.UserID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Cart created successfully",
        "cart":    cart,
    })
}

// GetCart retrieves a cart
func (ch *CartHandler) GetCart(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")
    if cartID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "cart id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    cart, err := ch.cartRepo.GetCart(ctx, cartID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, cart)
}

// AddItem adds an item to cart
func (ch *CartHandler) AddItem(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")
    if cartID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "cart id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    var req models.AddItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Verify cart exists
    cart, err := ch.cartRepo.GetCart(ctx, cartID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Create and add item
    item := models.NewCartItem(cartID, req.ProductID, req.Quantity, req.Price)
    if err := ch.cartRepo.AddItem(ctx, item); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to add item",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Update cart total
    newTotal := cart.Total + (req.Price * float64(req.Quantity))
    if err := ch.cartRepo.UpdateCartTotal(ctx, cartID, newTotal); err != nil {
        log.Printf("⚠️  Failed to update cart total: %v", err)
    }

    // Publish ItemAddedToCart event
    event := events.ItemAddedToCartEvent{
        BaseEvent: events.NewBaseEvent("ItemAddedToCart", cartID, "cart", ""),
        CartID:    cartID,
        ProductID: req.ProductID,
        Quantity:  req.Quantity,
        Price:     req.Price,
    }

    if err := ch.eventPublisher.PublishCartEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish ItemAddedToCart event: %v", err)
    }

    log.Printf("✓ Item added to cart: Product %d, Quantity %d", req.ProductID, req.Quantity)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Item added successfully",
        "item":    item,
    })
}

// RemoveItem removes an item from cart
func (ch *CartHandler) RemoveItem(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")
    productIDStr := c.Param("product_id")

    productID, err := strconv.ParseInt(productIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid product id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    if err := ch.cartRepo.RemoveItem(ctx, cartID, productID); err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "item not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    log.Printf("✓ Item removed from cart: Product %d", productID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Item removed successfully",
    })
}

// UpdateCart updates cart status
func (ch *CartHandler) UpdateCart(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")

    var req models.UpdateCartRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    if err := ch.cartRepo.UpdateCartStatus(ctx, cartID, req.Status); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to update cart",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ Cart updated: %s status → %s", cartID, req.Status)

    c.JSON(http.StatusOK, gin.H{
        "message": "Cart updated successfully",
    })
}

// DeleteCart deletes a cart
func (ch *CartHandler) DeleteCart(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")

    if err := ch.cartRepo.DeleteCart(ctx, cartID); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to delete cart",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ Cart deleted: %s", cartID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Cart deleted successfully",
    })
}

// CheckoutCart initiates checkout saga
func (ch *CartHandler) CheckoutCart(c *gin.Context) {
    ctx := context.Background()

    cartID := c.Param("id")

    var req models.CheckoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get cart
    cart, err := ch.cartRepo.GetCart(ctx, cartID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    if len(cart.Items) == 0 {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "cart is empty",
            Message: "cannot checkout empty cart",
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Create saga state
    correlationID := uuid.New().String()
    saga := models.NewSagaState(cartID, cart.UserID, correlationID)
    saga.Payload["order_id"] = req.OrderID
    saga.Payload["cart_id"] = cartID
    saga.Payload["user_id"] = cart.UserID
    saga.Payload["items"] = cart.Items
    saga.Payload["total"] = cart.Total

    if err := ch.sagaRepo.CreateSagaState(ctx, saga); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create saga state",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Update cart status
    if err := ch.cartRepo.UpdateCartStatus(ctx, cartID, "checked_out"); err != nil {
        log.Printf("⚠️  Failed to update cart status: %v", err)
    }

    // Publish CartCheckoutInitiated event (saga trigger)
    event := events.CartCheckoutInitiatedEvent{
        BaseEvent: events.NewBaseEvent("CartCheckoutInitiated", cartID, "cart", correlationID),
        CartID:    cartID,
        UserID:    cart.UserID,
        Total:     cart.Total,
    }

    if err := ch.eventPublisher.PublishCartEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish CartCheckoutInitiated event: %v", err)
    }

    log.Printf("✓ Checkout initiated: Cart %s, Correlation %s", cartID, correlationID)

    c.JSON(http.StatusAccepted, gin.H{
        "message":        "Checkout initiated",
        "correlation_id": correlationID,
        "saga_state":     saga,
    })
}