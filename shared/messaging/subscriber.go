package messaging

import (
	"encoding/json"
    "fmt"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sanketh-sg/prost/shared/events"
)

// MessageHandler defines the handler function for consuming messages
type MessageHandler func(message []byte) error

// Subscriber subscribes to events from RabbitMQ
type Subscriber struct {
	ch *amqp.Channel
	queueName string
}

// NewSubscriber creates a new event subscriber
func NewSubscriber(conn *Connection, queueName string) *Subscriber {
	return &Subscriber{
		ch: conn.GetChannel(),
		queueName: queueName,
	}
}

// Subscribe starts consuming messages from a queue
func (s *Subscriber) Subscribe(handler MessageHandler) error {
    deliveries, err := s.ch.Consume(
        s.queueName, // queue
        "",          // consumer
        false,       // auto-ack (we'll manually acknowledge)
        false,       // exclusive
        false,       // no-local
        false,       // no-wait
        nil,         // args
    )

    if err != nil {
        return fmt.Errorf("failed to consume from queue %s: %w", s.queueName, err)
    }

    log.Printf("Listening on queue: %s", s.queueName)

    // Process incoming messages
    for delivery := range deliveries {
        log.Printf(" Message received from %s", s.queueName)

        // Call the handler
        err := handler(delivery.Body)

        if err != nil {
            log.Printf(" Handler error: %v. Sending to DLQ...", err)
            // Negative acknowledgement sends to DLQ
            delivery.Nack(false, false) // don't requeue, go to DLQ
        } else {
            // Acknowledge successful processing
            delivery.Ack(false)
            log.Printf(" Message processed and acknowledged")
        }
    }

    return nil
}

// SubscribeWithRetry subscribes with automatic retry logic
func (s *Subscriber) SubscribeWithRetry(handler MessageHandler, maxRetries int) error {
	deliveries, err := s.ch.Consume(
		s.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to consume from queue: %s: %w", s.queueName, err)
	}

	for delivery := range deliveries{
		log.Printf(" Message received from %s", s.queueName)

		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			lastErr = handler(delivery.Body)
			if lastErr == nil {
				break
			}
			if attempt < maxRetries {
				log.Printf(" Attempt %d failed: %v. Retrying...", attempt, lastErr)
                time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			}
		}

		if lastErr != nil {
			log.Printf("All %d attempts failed: %v. Sending to DLQ...", maxRetries, lastErr)
			delivery.Nack(false,false)
		} else {
			delivery.Ack(false)
			log.Printf("Message delivered successfully")
		}
	}
	return nil
}


// ParseEvent parses JSON message into an event
func (s *Subscriber) ParseEvent(data []byte) (interface{}, error) {
    var baseEvent struct {
        EventType string `json:"event_type"`
    }

    if err := json.Unmarshal(data, &baseEvent); err != nil {
        return nil, fmt.Errorf("failed to unmarshal base event: %w", err)
    }

    return events.UnmarshalEvent(data, baseEvent.EventType)
}