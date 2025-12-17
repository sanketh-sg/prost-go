package handlers

import (
    "context"
    "log"
	"fmt"

    "github.com/sanketh-sg/prost/services/products/models"
    "github.com/sanketh-sg/prost/services/products/repository"
    "github.com/sanketh-sg/prost/shared/events"
)

// handleStockReserved processes StockReservedEvent from RabbitMQ
// Why: When Cart or Order service reserves inventory, we record it
// This creates an inventory_reservations record linking the product to the order
func HandleStockReserved(ctx context.Context, event *events.StockReservedEvent, inventoryRepo *repository.InventoryReservationRepository) error {
    log.Printf("✓ RabbitMQ: Processing StockReservedEvent: ProductID=%d, Quantity=%d, ReservationID=%s", 
        event.ProductID, event.Quantity, event.ReservationID)
    
    reservation := &models.InventoryReservation{
        ProductID:     event.ProductID,
        Quantity:      event.Quantity,
        OrderID:       event.OrderID,
        ReservationID: event.ReservationID,
        Status:        "reserved",
    }
    
    if err := inventoryRepo.CreateReservation(ctx, reservation); err != nil {
        log.Printf("❌ Failed to create reservation: %v", err)
        return fmt.Errorf("failed to create reservation: %w", err)
    }
    
    log.Printf("✓ Reservation created from RabbitMQ event: %s", event.ReservationID)
    return nil
}

// handleStockReleased processes StockReleasedEvent from RabbitMQ (compensation)
// Why: When an order fails or is cancelled, we release the reserved inventory
// This marks the reservation as "released" so it's no longer counted as reserved
func HandleStockReleased(ctx context.Context, event *events.StockReleasedEvent, inventoryRepo *repository.InventoryReservationRepository) error {
    log.Printf("✓ RabbitMQ: Processing StockReleasedEvent: ProductID=%d, Quantity=%d, Reason=%s", 
        event.ProductID, event.Quantity, event.Reason)
    
    if err := inventoryRepo.ReleaseReservation(ctx, event.ReservationID); err != nil {
        log.Printf("❌ Failed to release reservation: %v", err)
        return fmt.Errorf("failed to release reservation: %w", err)
    }
    
    log.Printf("✓ Reservation released from RabbitMQ event: %s (Reason: %s)", event.ReservationID, event.Reason)
    return nil
}
