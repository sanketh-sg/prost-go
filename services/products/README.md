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

Context is used for timeout control and cancellation. QueryRowContext allows the database query to be stopped if:

Request timeout exceeded - If the HTTP client disconnects or the request deadline passes, the context triggers and cancels the query
Custom timeout - You can set a specific timeout with context.WithTimeout() to stop slow queries
Cancellation signal - Parent operations can cancel queries with context.WithCancel()
Without context, a slow database query could run forever, blocking resources and hurting performance. The context ensures queries respect timeouts and don't waste resources on requests that have already been abandoned by the client.

Browser Request: GET /users/123
    ↓
Server: SELECT * FROM users WHERE id = 123
    ↓
Database: Slow query (taking 30 seconds)
    ↓
User gets impatient (after 5 seconds):
    ├─ Closes browser tab
    └─ Connection closed

Server status:
❌ Still waiting for database (no context)
    → Wasted resources, CPU still running
    → Other requests can't use this connection
    
✅ With context timeout (5 seconds)
    → Query cancelled automatically
    → Connection freed
    → Can serve other requests


Different operation types get different timeouts:

Fast reads: 1-2 seconds
Writes: 3-5 seconds
Complex operations: 5-10 seconds
The timeout should always be less than the HTTP request timeout to fail fast.

The handler receives *gin.Context as a parameter, which is passed by the Gin framework when the HTTP request arrives. Inside the handler, I extract the Go context.Context from the Gin context using c.Request.Context(). This preserves the request context—including user information from JWT, request deadlines, and request IDs for tracing.

Gin context: c *gin.Context
├─ Request ID
├─ User ID (from JWT)
├─ Request deadline
├─ Request path
└─ ... other HTTP context

used for request data extraction, response writing, context propagation, insert custom values.

Produces:
products.events (Topic Exchange)
├─ product.stock.reserved  → StockReservedEvent
└─ product.stock.released  → StockReleasedEvent

Consumes:
orders.events (Topic Exchange)  → products.events.queue
├─ order.confirmed  → OrderConfirmedEvent
├─ order.failed     → OrderFailedEvent
└─ order.cancelled  → OrderCancelledEvent

