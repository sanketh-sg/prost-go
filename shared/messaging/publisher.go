package messaging

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sanketh-sg/prost/shared/events"
)

type Publisher struct {
	ch *amqp.Channel
	exchange string
}

func NewPublisher(conn *Connection, exchange string) *Publisher {
	return &Publisher{
		ch: conn.ch,
		exchange: exchange,
	}
}

func (pub *Publisher) PublishEvent(ctx context.Context, event interface{}, routingKey string) error {
	// Marshal event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = pub.ch.PublishWithContext(
		ctx,
		pub.exchange,
		routingKey,
		false, //mandatory
		false, //immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
			Timestamp: time.Now(),
			DeliveryMode: amqp.Persistent, //Message persists if RabbitMQ restarts
		},
	)

	if err != nil {
        return fmt.Errorf("failed to publish event: %w", err)
    }

    var eventID string
    if baseEvent, ok := event.(interface{ GetEventID() string }); ok {
        eventID = baseEvent.GetEventID()
    }

    log.Printf("Event published: %s (routing key: %s, event_id: %s)", pub.exchange, routingKey, eventID)
    return nil
}

func (pub *Publisher) PublishProductEvent(ctx context.Context, event interface{}) error {
	var routingKey string

	switch event.(type) { //The switch itself performs the type comparison internally.
	case events.ProductCreatedEvent: routingKey = "product.created"
	case events.ProductUpdatedEvent: routingKey = "product.updated"
	case events.StockReservedEvent: routingKey = "product.stock.reserved"
	case events.StockReleasedEvent: routingKey = "product.stock.released"
	default:
		return fmt.Errorf("unknown product event type: %T", event)
	}

	return pub.PublishEvent(ctx, event,routingKey)
}


func (p *Publisher) PublishOrderEvent(ctx context.Context, event interface{}) error {
    var routingKey string

    switch event.(type) {
    case events.OrderPlacedEvent:
        routingKey = "order.placed"
    case events.OrderConfirmedEvent:
        routingKey = "order.confirmed"
    case events.OrderFailedEvent:
        routingKey = "order.failed"
    case events.OrderCancelledEvent:
        routingKey = "order.cancelled"
    case events.OrderShippedEvent:
        routingKey = "order.shipped"
    default:
        return fmt.Errorf("unknown order event type: %T", event)
    }

    return p.PublishEvent(ctx, event, routingKey)
}