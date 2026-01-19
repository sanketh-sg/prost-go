package handlers

import (
	"context"
	"fmt"
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
	sharedModels "github.com/sanketh-sg/prost/shared/models"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	cartRepo          *repository.CartRepository
	sagaRepo          *repository.SagaStateRepository
	inventoryLockRepo *repository.InventoryLockRepository
	idempotencyStore  *db.IdempotencyStore
	eventPublisher    *messaging.Publisher
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

// Helper: GetUserIDFromContext extracts userID from auth middleware
func (ch *CartHandler) getUserIDFromContext(c *gin.Context) (string, error) {
    userID, exists := c.Get("user_id")
    if !exists {
        return "", fmt.Errorf("user_id not found in context (missing auth?)")
    }

    userIDStr, ok := userID.(string)
    if !ok {
        return "", fmt.Errorf("user_id is not a string")
    }

    return userIDStr, nil
}

// CreateCart gets user's active cart or creates new one
func (ch *CartHandler) CreateCart(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "unauthorized",
            Message: err.Error(),
            Code:    http.StatusUnauthorized,   
        })
        return
    }

    // Get existing active cart
    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err == nil && cart != nil {
        log.Printf("✓ Returning existing cart: %s for user %s", cart.ID, userID)
        c.JSON(http.StatusOK, gin.H{
            "message": "Cart retrieved successfully",
            "cart":    cart,
        })
        return
    }

    // Create new cart if none exists
    log.Println("Cart not found creating a new cart...")
    newCart := models.NewCart(userID)
    if err := ch.cartRepo.CreateCart(ctx, newCart); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create cart",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("New cart created: %s for user %s", newCart.ID, userID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Cart created successfully",
        "cart":    newCart,
    })
}

// GetCart retrieves user's active cart
func (ch *CartHandler) GetCart(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "unauthorized",
            Message: err.Error(),
            Code:    http.StatusUnauthorized,   
        })
        return
    }

    // Get existing active cart
    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err != nil || cart == nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: "No active cart exists for this user",
            Code:    http.StatusNotFound,
        })
        if err != nil {
            log.Printf("Error retrieving cart for user %s: %v", userID, err)
        }
        return
    }

    log.Printf("✓ Cart retrieved: %s for user %s", cart.ID, userID)
    c.JSON(http.StatusOK, gin.H{
        "message": "Cart retrieved successfully",
        "cart":    cart,
    })
}


// AddItem adds an item to user's cart
func (ch *CartHandler) AddItem(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "unauthorized",
            Message: err.Error(),
            Code:    http.StatusUnauthorized,
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

    // Get user's active cart
    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err != nil || cart == nil {
      newCart := models.NewCart(userID)
        if err := ch.cartRepo.CreateCart(ctx, newCart); err != nil {
            c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                Error:   "failed to create cart",
                Message: err.Error(),
                Code:    http.StatusInternalServerError,
            })
            return
        }
        cart = newCart
        log.Printf("✓ New cart created for user %s: %s", userID, cart.ID)
    }


    // Create and add item
    item := models.NewCartItem(cart.ID, req.ProductID, req.Quantity, req.Price)
    if err := ch.cartRepo.AddItem(ctx, item); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to add item",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Update cart total using helper
    if err := ch.updateCartTotal(ctx, cart.ID); err != nil {
        log.Printf("Failed to update cart total: %v", err)
    }

    // Get updated cart for response
    updatedCart, _ := ch.cartRepo.GetCart(ctx, cart.ID)

    log.Printf("✓ Item added to cart: Product %d, Quantity %d", req.ProductID, req.Quantity)

    c.JSON(http.StatusCreated, gin.H{
        "message":   "Item added successfully",
        "item":      item,
        "new_total": updatedCart.Total,
    })
}

// RemoveItem removes an item from cart
func (ch *CartHandler) RemoveItem(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "unauthorized",
            Message: err.Error(),
            Code:    http.StatusUnauthorized,
        })
        return
    }

    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

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

    // Find the item being removed to get its quantity
    var itemQuantity int
    itemFound := false
    for _, item := range cart.Items {
        if item.ProductID == productID {
            itemQuantity = item.Quantity
            itemFound = true
            break
        }
    }

    
    // Validate item exists before removing
    if !itemFound {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "item not found",
            Message: "product not in cart",
            Code:    http.StatusNotFound,
        })
        return
    }

    // Remove item from cart
    if err := ch.cartRepo.RemoveItem(ctx, cart.ID, productID); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to remove item",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    if err := ch.updateCartTotal(ctx, cart.ID); err != nil {
        log.Printf("⚠️  Failed to update cart total: %v", err)
    }


    // Get updated cart for response
    updatedCart, _ := ch.cartRepo.GetCart(ctx, cart.ID)


    log.Printf("Item removed from cart: Product %d, Quantity %d, New Total: %.2f", productID, itemQuantity, updatedCart.Total)

    c.JSON(http.StatusOK, gin.H{
        "message":   "Item removed successfully",
        "new_total": updatedCart.Total,
    })
}

// updateCartTotal recalculates and updates cart total based on current items
// Why: Centralizes total calculation logic, prevents inconsistencies
func (ch *CartHandler) updateCartTotal(ctx context.Context, cartID string) error {
    cart, err := ch.cartRepo.GetCart(ctx, cartID)
    if err != nil {
        return fmt.Errorf("failed to get cart: %w", err)
    }

    // Recalculate total from all items
    newTotal := 0.0
    for _, item := range cart.Items {
        newTotal += item.Price * float64(item.Quantity)
    }

    // Update in database
    if err := ch.cartRepo.UpdateCartTotal(ctx, cartID, newTotal); err != nil {
        return fmt.Errorf("failed to update cart total: %w", err)
    }

    log.Printf("✓ Cart total updated: %s → %.2f", cartID, newTotal)
    return nil
}

// DeleteCart deletes a cart
func (ch *CartHandler) DeleteCart(c *gin.Context) {
	// ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

    userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error: "User not found, Unauthorised",
            Message: err.Error(),
            Code: http.StatusUnauthorized,
        })
        return
    }
    
    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error: "cart not found",
            Message: err.Error(),
            Code: http.StatusNotFound,
        })
        return
    }

	if err := ch.cartRepo.DeleteCart(ctx, cart.ID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "failed to delete cart",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	log.Printf("Cart deleted: %s", cart.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart deleted successfully",
	})
}

// CheckoutCart initiates checkout saga
func (ch *CartHandler) CheckoutCart(c *gin.Context) {
	// ctx := context.Background()
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID, err := ch.getUserIDFromContext(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "unauthorized",
            Message: err.Error(),
            Code:    http.StatusUnauthorized,
        })
        return
    }

    cart, err := ch.cartRepo.GetCartByUserID(ctx, userID)
    if err != nil || cart == nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "cart not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
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
	saga := models.NewSagaState(cart.ID, userID, correlationID)
	saga.Payload["order_id"] = req.OrderID
	saga.Payload["cart_id"] = cart.ID
	saga.Payload["user_id"] = userID
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
	if err := ch.cartRepo.UpdateCartStatus(ctx, cart.ID, "checked_out"); err != nil {
		log.Printf("⚠️  Failed to update cart status: %v", err)
	}

	// Publish CartCheckoutInitiated event (saga trigger)
	event := events.CartCheckoutInitiatedEvent{
		BaseEvent: events.NewBaseEvent("CartCheckoutInitiated", cart.ID, "cart", correlationID),
		CartID:    cart.ID,
		UserID:    cart.UserID,
		Total:     cart.Total,
		Items:      ch.convertCartItemsToOrderItems(cart.Items),
	}

	if err := ch.eventPublisher.PublishCartEvent(ctx, event); err != nil {
		log.Printf("⚠️  Failed to publish CartCheckoutInitiated event: %v", err)
	}

	log.Printf("✓ Checkout initiated: Cart %s, Correlation %s", cart.ID, correlationID)

	c.JSON(http.StatusAccepted, gin.H{
		"message":        "Checkout initiated",
		"correlation_id": correlationID,
		"saga_state":     saga,
	})
}

func (ch *CartHandler) convertCartItemsToOrderItems(cartItems []models.CartItem) []sharedModels.OrderItem{
    orderItems := make([]sharedModels.OrderItem, len(cartItems))
    for i, cartItem := range cartItems {
        orderItems[i] = sharedModels.OrderItem{
            ProductID: cartItem.ProductID,
            Quantity: cartItem.Quantity,
            Price: cartItem.Price,
        }
    }
    return orderItems
}