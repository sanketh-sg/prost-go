1. Setup env variables
2. Enable/Disable gin mode to production
3. Establish DB connection 
4. Establish RabbitMQ channel -> request for new conn, establish a channel using the conn return a channel, conn
5. Get Topology and set it up
6. Initialize repos for CRUD ops
7. Init pub & sub















┌──────────────────────────────────────────────────────────────┐
│                         EXCHANGES                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐     │
│  │  products.events (Topic Exchange)                   │     │
│  │  Type: topic                                        │     │
│  │  Durable: true                                      │     │
│  │  Purpose: Publish product events                    │     │
│  └─────────────────────────────────────────────────────┘    `│
│           ↓                                                  │
│           │ (routes messages with routing key)               │
│           │                                                  │
│  ┌─────────────────────────────────────────────────────┐     │
│  │  products.events.dlx (Dead Letter Exchange)         │     │
│  │  Type: (implicit)                                   │     │
│  │  Purpose: Handle failed/expired messages            │     │
│  └─────────────────────────────────────────────────────┘     │ 
│                                                              │
└──────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────┐
│                       QUEUES                                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Main Queue:                                                 │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  products.events.queue                               │    │
│  │  Durable: true                                       │    │
│  │  TTL: 24 hours (86400000 ms)                         │    │
│  │  DLX: products.events.dlx                            │    │
│  │  Purpose: Consume product events                     │    │
│  └──────────────────────────────────────────────────────┘    │
│           ↑                                                  │
│           │ (consumes messages)                              │
│           │                                                  │
│  Dead Letter Queue:                                          │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  products.events.dlq                                 │    │
│  │  Durable: true                                       │    │
│  │  Purpose: Store failed/expired messages              │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              |
└──────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────┐
│                     BINDINGS                                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Binding 1: Main Queue                                       │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Exchange: products.events                           │    │
│  │  Queue:    products.events.queue                     │    │
│  │  Routing Key: product.*                              │    │
│  │  Effect: Queue receives all product.* events         │    │
│  └──────────────────────────────────────────────────────┘    │ 
│                                                              │
│  Binding 2: Dead Letter Queue                                │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Exchange: products.events.dlx                       │    │
│  │  Queue:    products.events.dlq                       │    │
│  │  Routing Key: #  (all messages)                      │    │
│  │  Effect: DLX forwards all failed msgs to DLQ         │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                              │
└──────────────────────────────────────────────────────────────┘