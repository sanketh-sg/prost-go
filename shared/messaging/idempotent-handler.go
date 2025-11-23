package messaging

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

	"github.com/sanketh-sg/prost/shared/db"
)

// IdempotentHandler wraps a MessageHandler with idempotency checking
type IdempotentHandler struct {
	idempotencyStore *db.IdempotencyStore
	serviceName string
	handler MessageHandler
}

// NewIdempotentHandler creates a new idempotent handler
func NewIdempotentHandler(idempotencyStore *db.IdempotencyStore, serviceName string, handler MessageHandler) *IdempotentHandler{
	return &IdempotentHandler{
		idempotencyStore: idempotencyStore,
		serviceName: serviceName,
		handler: handler,
	}
} 

// Handle processes a message with idempotency checking
func (ih *IdempotentHandler) Handle(ctx context.Context, data []byte) error {
    // Extract event_id from message
    var baseEvent struct {
        EventID string `json:"event_id"`
    }

    if err := json.Unmarshal(data, &baseEvent); err != nil {
        return fmt.Errorf("failed to unmarshal event_id: %w", err)
    }

    eventID := baseEvent.EventID
    if eventID == "" {
        return fmt.Errorf("event_id is required for idempotency")
    }

    // Check if already processed
    processed, err := ih.idempotencyStore.IsProcessed(ctx, eventID, ih.serviceName)
    if err != nil {
        log.Printf("⚠️  Failed to check idempotency: %v", err)
        // Continue anyway - better to process twice than not at all
    }

    if processed {
        log.Printf("⏭️  Event %s already processed. Skipping...", eventID)
        return nil // Don't error, just skip
    }

    // Process the message
    err = ih.handler(data)

    // Record processing result
    result := "success"
    if err != nil {
        result = "failed"
    }

    if recordErr := ih.idempotencyStore.RecordProcessed(ctx, eventID, ih.serviceName, "process_event", result); recordErr != nil {
        log.Printf("⚠️  Failed to record idempotency: %v", recordErr)
    }

    return err
}
