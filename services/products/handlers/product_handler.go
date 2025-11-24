package handlers

import (
    "context"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/products/models"
    "github.com/sanketh-sg/prost/services/products/repository"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/events"
    "github.com/sanketh-sg/prost/shared/messaging"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
    productRepo     *repository.ProductRepository
    categoryRepo    *repository.CategoryRepository
    inventoryRepo   *repository.InventoryReservationRepository
    idempotencyStore *db.IdempotencyStore
    eventPublisher  *messaging.Publisher
}

// NewProductHandler creates new product handler
func NewProductHandler(
    productRepo *repository.ProductRepository,
    categoryRepo *repository.CategoryRepository,
    inventoryRepo *repository.InventoryReservationRepository,
    idempotencyStore *db.IdempotencyStore,
    eventPublisher *messaging.Publisher,
) *ProductHandler {
    return &ProductHandler{
        productRepo:      productRepo,
        categoryRepo:     categoryRepo,
        inventoryRepo:    inventoryRepo,
        idempotencyStore: idempotencyStore,
        eventPublisher:   eventPublisher,
    }
}

// Health handles health check
func (ph *ProductHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "products",
        "time":    time.Now().UTC(),
    })
}

// CreateCategory creates a new category
func (ph *ProductHandler) CreateCategory(c *gin.Context) {
    ctx := context.Background()

    var req models.CreateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    category := models.NewCategory(req.Name, req.Description)
    if err := ph.categoryRepo.CreateCategory(ctx, category); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create category",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ Category created: %s (ID: %d)", category.Name, category.ID)

    c.JSON(http.StatusCreated, gin.H{
        "message":  "Category created successfully",
        "category": category,
    })
}

// GetCategory retrieves a category
func (ph *ProductHandler) GetCategory(c *gin.Context) {
    ctx := context.Background()

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid category id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    category, err := ph.categoryRepo.GetCategory(ctx, id)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "category not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, category)
}

// GetCategories retrieves all categories
func (ph *ProductHandler) GetCategories(c *gin.Context) {
    ctx := context.Background()

    categories, err := ph.categoryRepo.GetAllCategories(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to get categories",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "categories": categories,
        "count":      len(categories),
    })
}

// CreateProduct creates a new product
func (ph *ProductHandler) CreateProduct(c *gin.Context) {
    ctx := context.Background()

    var req models.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    product := models.NewProduct(
        req.Name,
        req.Description,
        req.Price,
        req.SKU,
        req.CategoryID,
        req.Stock,
        req.ImageURL,
    )

    if err := ph.productRepo.CreateProduct(ctx, product); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create product",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Publish ProductCreated event
    event := events.ProductCreatedEvent{
        BaseEvent: events.NewBaseEvent("ProductCreated", strconv.FormatInt(product.ID, 10), "product", ""),
        Name:      product.Name,
        Description: product.Description,
        Price:     product.Price,
        SKU:       product.SKU,
        CategoryID: product.CategoryID,
        ImageURL:  product.ImageURL,
    }

    if err := ph.eventPublisher.PublishProductEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish ProductCreated event: %v", err)
    }

    log.Printf("✓ Product created: %s (ID: %d)", product.Name, product.ID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Product created successfully",
        "product": product,
    })
}

// GetProduct retrieves a product
func (ph *ProductHandler) GetProduct(c *gin.Context) {
    ctx := context.Background()

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid product id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    product, err := ph.productRepo.GetProduct(ctx, id)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "product not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, product)
}

// GetProducts retrieves all products
func (ph *ProductHandler) GetProducts(c *gin.Context) {
    ctx := context.Background()

    var categoryID *int64
    if catID := c.Query("category_id"); catID != "" {
        id, err := strconv.ParseInt(catID, 10, 64)
        if err == nil {
            categoryID = &id
        }
    }

    products, err := ph.productRepo.GetAllProducts(ctx, categoryID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to get products",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "products": products,
        "count":    len(products),
    })
}

// UpdateProduct updates a product
func (ph *ProductHandler) UpdateProduct(c *gin.Context) {
    ctx := context.Background()

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid product id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    var req models.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    product, err := ph.productRepo.GetProduct(ctx, id)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "product not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Update fields if provided
    if req.Name != "" {
        product.Name = req.Name
    }
    if req.Description != "" {
        product.Description = req.Description
    }
    if req.Price > 0 {
        product.Price = req.Price
    }
    if req.Stock >= 0 {
        product.StockQuantity = req.Stock
    }
    if req.ImageURL != "" {
        product.ImageURL = req.ImageURL
    }

    if err := ph.productRepo.UpdateProduct(ctx, product); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to update product",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Publish ProductUpdated event
    event := events.ProductUpdatedEvent{
        BaseEvent: events.NewBaseEvent("ProductUpdated", strconv.FormatInt(product.ID, 10), "product", ""),
        Name:      product.Name,
        Description: product.Description,
        Price:     product.Price,
        ImageURL:  product.ImageURL,
    }

    if err := ph.eventPublisher.PublishProductEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish ProductUpdated event: %v", err)
    }

    log.Printf("✓ Product updated: %s (ID: %d)", product.Name, product.ID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Product updated successfully",
        "product": product,
    })
}

// DeleteProduct deletes a product
func (ph *ProductHandler) DeleteProduct(c *gin.Context) {
    ctx := context.Background()

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid product id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    if err := ph.productRepo.DeleteProduct(ctx, id); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to delete product",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ Product deleted: ID: %d", id)

    c.JSON(http.StatusOK, gin.H{
        "message": "Product deleted successfully",
    })
}

// GetInventory gets current inventory for a product
func (ph *ProductHandler) GetInventory(c *gin.Context) {
    ctx := context.Background()

    productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid product id",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    product, err := ph.productRepo.GetProduct(ctx, productID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "product not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    reserved, err := ph.inventoryRepo.GetProductReservations(ctx, productID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to get reservations",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    available := product.StockQuantity - reserved

    c.JSON(http.StatusOK, gin.H{
        "product_id": productID,
        "total_stock": product.StockQuantity,
        "reserved": reserved,
        "available": available,
    })
}

// ReserveInventory reserves inventory for an order
func (ph *ProductHandler) ReserveInventory(c *gin.Context) {
    ctx := context.Background()

    var req models.ReserveInventoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Check product exists and has stock
    product, err := ph.productRepo.GetProduct(ctx, req.ProductID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "product not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Check available stock
    reserved, err := ph.inventoryRepo.GetProductReservations(ctx, req.ProductID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to check reservations",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    available := product.StockQuantity - reserved
    if available < req.Quantity {
        c.JSON(http.StatusConflict, models.ErrorResponse{
            Error:   "insufficient stock",
            Message: "not enough inventory available",
            Code:    http.StatusConflict,
        })
        return
    }

    // Create reservation
    reservation := models.NewInventoryReservation(req.ProductID, req.Quantity, req.OrderID, req.ReservationID)
    if err := ph.inventoryRepo.CreateReservation(ctx, reservation); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to reserve inventory",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Publish StockReserved event
    event := events.StockReservedEvent{
        BaseEvent: events.NewBaseEvent("StockReserved", strconv.FormatInt(req.ProductID, 10), "product", strconv.FormatInt(req.OrderID, 10)),
        ProductID: req.ProductID,
        Quantity: req.Quantity,
        OrderID: req.OrderID,
        ReservationID: req.ReservationID,
    }

    if err := ph.eventPublisher.PublishProductEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish StockReserved event: %v", err)
    }

    log.Printf("✓ Stock reserved: Product %d, Quantity %d, Order %d", req.ProductID, req.Quantity, req.OrderID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "Inventory reserved successfully",
        "reservation": reservation,
    })
}

// ReleaseInventory releases a reserved inventory
func (ph *ProductHandler) ReleaseInventory(c *gin.Context) {
    ctx := context.Background()

    var req models.ReleaseInventoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get reservation to find product ID
    reservation, err := ph.inventoryRepo.GetReservation(ctx, req.ReservationID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "reservation not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Release reservation
    if err := ph.inventoryRepo.ReleaseReservation(ctx, req.ReservationID); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to release inventory",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Publish StockReleased event
    event := events.StockReleasedEvent{
        BaseEvent: events.NewBaseEvent("StockReleased", strconv.FormatInt(reservation.ProductID, 10), "product", strconv.FormatInt(reservation.OrderID, 10)),
        ProductID: reservation.ProductID,
        Quantity: reservation.Quantity,
        ReservationID: req.ReservationID,
        Reason: req.Reason,
    }

    if err := ph.eventPublisher.PublishProductEvent(ctx, event); err != nil {
        log.Printf("⚠️  Failed to publish StockReleased event: %v", err)
    }

    log.Printf("✓ Stock released: Product %d, Reservation %s, Reason: %s", reservation.ProductID, req.ReservationID, req.Reason)

    c.JSON(http.StatusOK, gin.H{
        "message": "Inventory released successfully",
    })
}