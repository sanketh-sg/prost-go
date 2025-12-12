package messaging

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn *amqp.Connection
	ch *amqp.Channel
}

func NewRmqConnection(connURL string)(*Connection, error){
	
	var conn *amqp.Connection
	var err error

	log.Println(connURL)

	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(connURL)

		if err == nil {
			break
		}

		if i<4 {
			log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v. Retrying in 2s...", i+1, err)
			time.Sleep(2 * time.Second)
		}	
	}
	if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %w", err)
    }

	ch, err := conn.Channel()	
	if err != nil {
		conn.Close()
        return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("Connected to RabbitMQ...:)")

	return &Connection{
		conn: conn,
		ch: ch,
		},nil
}

// Close closes the RabbitMQ connection
func (c *Connection) Close() error {
    if err := c.ch.Close(); err != nil {
        return fmt.Errorf("failed to close channel: %w", err)
    }

    if err := c.conn.Close(); err != nil {
        return fmt.Errorf("failed to close connection: %w", err)
    }

    return nil
}

// GetChannel returns the AMQP channel
func (conn *Connection) GetChannel() *amqp.Channel {
    return conn.ch
}

func (conn *Connection) SetupRabbitMQ(topology MessagingTopology) error{
	log.Println("Setting up RabbitMQ topology...")

	// setup QoS
	if err:= conn.ch.Qos(1,0,false); err != nil {
		return fmt.Errorf("failed to setup QoS: %w", err)
	}

	// create exchange

	for _, ex := range topology.Exchanges { //_ represents the index which we dont need blank identifier
		if err := conn.ch.ExchangeDeclare(
			ex.Name,
			ex.Type,
			ex.Durable,
			ex.AutoDelete,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
		log.Printf("Exchange created: %s (%s)", ex.Name, ex.Type)
	}
	
	//DLX Dead Letter Exchange
	dlxExchanges := []string{
		"products.events.dlx",
        "cart.events.dlx",
        "orders.events.dlx",
	}

	for _, dlx := range dlxExchanges {
		err := conn.ch.ExchangeDeclare(
			dlx,
			"topic",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare DLX %s: %w", dlx, err)
		}
		 log.Printf("DLX created: %s", dlx)
	}

	// Create Queues

	for _, queueObj := range topology.Queues{
		_, err := conn.ch.QueueDeclare(
			queueObj.Name,
			queueObj.Durable,
			queueObj.AutoDelete,
			queueObj.Exclusive,
			queueObj.NoWait,
			amqp.Table(queueObj.Arguments),
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueObj.Name, err)
		}
		log.Printf("Queue created: %s", queueObj.Name)
	}

	for _, bindObj := range topology.Bindings{
		err := conn.ch.QueueBind(
			bindObj.QueueName,
			bindObj.RoutingKey,
			bindObj.ExchangeName,
			bindObj.NoWait,
			amqp.Table(bindObj.Arguments),
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", bindObj.QueueName, bindObj.ExchangeName, err)
        }
        log.Printf(" Binding: %s -> %s (key: %s)", bindObj.ExchangeName, bindObj.QueueName, bindObj.RoutingKey)
	}
	    log.Println("RabbitMQ topology setup complete")
    return nil
}

