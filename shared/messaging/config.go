package messaging

import (
	"log"
)

// ExchangeConfig defines an exchange configuration
type ExchangeConfig struct {
	Name       string
	Type       string // "topic", "direct", "fanout"
	Durable    bool
	AutoDelete bool
}

// QueueConfig defines a queue configuration
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  map[string]interface{}
}

// BindingConfig defines a queue-to-exchange binding
type BindingConfig struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Arguments    map[string]interface{}
}

// MessagingTopology defines the complete RabbitMQ setup
type MessagingTopology struct {
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Bindings  []BindingConfig
}

// GetProstTopology returns the messaging topology for Prost microservices
func GetProstTopology() MessagingTopology {
	return MessagingTopology{
		Exchanges: []ExchangeConfig{
			// ========== Main Event Exchanges ==========
			{
				Name:       "products.events",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},
			{
				Name:       "cart.events",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},
			{
				Name:       "orders.events",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},

			// ========== Dead Letter Exchanges ==========
			{
				Name:       "products.events.dlx",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},
			{
				Name:       "cart.events.dlx",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},
			{
				Name:       "orders.events.dlx",
				Type:       "topic",
				Durable:    true,
				AutoDelete: false,
			},
		},
		Queues: []QueueConfig{
			// Products service queues
			{
				Name:       "products.events.queue",
				Durable:    true,
				AutoDelete: false,
				Arguments: map[string]interface{}{
					"x-dead-letter-exchange": "products.events.dlx",
					"x-message-ttl":          86400000, // 24 hours
				},
			},
			{
				Name:       "products.events.dlq",
				Durable:    true,
				AutoDelete: false,
				Arguments:  map[string]interface{}{},
			},

			// Cart service queues
			{
				Name:       "cart.events.queue",
				Durable:    true,
				AutoDelete: false,
				Arguments: map[string]interface{}{
					"x-dead-letter-exchange": "cart.events.dlx",
					"x-message-ttl":          86400000,
				},
			},
			{
				Name:       "cart.events.dlq",
				Durable:    true,
				AutoDelete: false,
				Arguments:  map[string]interface{}{},
			},

			// Orders service queues
			{
				Name:       "orders.events.queue",
				Durable:    true,
				AutoDelete: false,
				Arguments: map[string]interface{}{
					"x-dead-letter-exchange": "orders.events.dlx",
					"x-message-ttl":          86400000,
				},
			},
			{
				Name:       "orders.events.dlq",
				Durable:    true,
				AutoDelete: false,
				Arguments:  map[string]interface{}{},
			},
		},
		Bindings: []BindingConfig{
			// Products service bindings
			{
				QueueName:    "products.events.queue",
				ExchangeName: "products.events",
				RoutingKey:   "product.*",
			},
			{
				QueueName:    "products.events.dlq",
				ExchangeName: "products.events.dlx",
				RoutingKey:   "#",
			},
			{
				QueueName:    "products.events.queue",
				ExchangeName: "orders.events",
				RoutingKey:   "order.*",
			},
			// Cart service bindings - listens to product and cart events
			{
				QueueName:    "cart.events.queue",
				ExchangeName: "products.events",
				RoutingKey:   "product.stock.*",
			},
			{
				QueueName:    "cart.events.queue",
				ExchangeName: "cart.events",
				RoutingKey:   "cart.*",
			},
			{
				QueueName:    "cart.events.dlq",
				ExchangeName: "cart.events.dlx",
				RoutingKey:   "#",
			},
			{
				QueueName:    "cart.events.queue",
				ExchangeName: "orders.events",
				RoutingKey:   "order.failed",
			},
			// Orders service bindings - listens to cart and order events
			{
				QueueName:    "orders.events.queue",
				ExchangeName: "cart.events",
				RoutingKey:   "cart.checkout.*",
			},
			{
				QueueName:    "orders.events.queue",
				ExchangeName: "orders.events",
				RoutingKey:   "order.*",
			},
			{
				QueueName:    "orders.events.dlq",
				ExchangeName: "orders.events.dlx",
				RoutingKey:   "#",
			},
		},
	}
}

// Log prints the topology configuration
func (t MessagingTopology) Log() {
	log.Println("\n=== RabbitMQ Topology Configuration ===")

	log.Println("\nExchanges:")
	for _, ex := range t.Exchanges {
		log.Printf("  - %s (%s, durable: %v)", ex.Name, ex.Type, ex.Durable)
	}

	log.Println("\nQueues:")
	for _, q := range t.Queues {
		log.Printf("  - %s (durable: %v, ttl: %v)", q.Name, q.Durable, q.Arguments["x-message-ttl"])
	}

	log.Println("\nBindings:")
	for _, b := range t.Bindings {
		log.Printf("  - %s -> %s (routing key: %s)", b.ExchangeName, b.QueueName, b.RoutingKey)
	}

	log.Println("\n========================================")
}
