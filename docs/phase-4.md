## Phase 4: Products Service - Key Learnings Summary

### 1. **Event Publishing Architecture**

```
When product action occurs:
1. Handler processes HTTP request
2. Repository updates database
3. Event is created with event_id (UUID)
4. Publisher sends to RabbitMQ exchange
5. Event persists in queue for subscribers

Example: CreateProduct flow
POST /products {name, price, sku, ...}
  ↓
Handler validates → Repository inserts → Product created
  ↓
ProductCreatedEvent {
  event_id: "uuid-123" (for idempotency)
  event_type: "ProductCreated"
  aggregate_id: "product-5" (product ID)
  aggregate_type: "product"
  correlation_id: "" (empty at source)
  timestamp: now()
  data: {name, price, sku, ...}
}
  ↓
Publisher → products.events exchange → products.events.queue
  ↓
Other services subscribe and react
```

**Key Learning:** Events are the "notifications" that trigger async workflows. Every state change = event published.

---

### 2. **Inventory Reservation Pattern (Critical for Sagas)**

```
Three-tier inventory tracking:

1. PRODUCT STOCK (catalog.products.stock_quantity)
   ├─ Physical inventory you own
   └─ Decremented only when order ships
   
2. ACTIVE RESERVATIONS (catalog.inventory_reservations)
   ├─ Quantity locked for pending orders
   ├─ TTL: 24 hours (auto-expire stale reservations)
   └─ Status: reserved → released
   
3. AVAILABLE = STOCK - ACTIVE_RESERVATIONS
   ├─ What customers can actually order now
   └─ Decreases as orders come in
   └─ Increases as reservations expire/release

Scenario:
- Total stock: 100 laptops
- Active reservations: 30 laptops (pending orders)
- Available: 70 laptops (can sell)

If customer buys 5:
- Check: 70 ≥ 5? Yes
- Create reservation: reservation_id = "res-uuid"
- Publish: StockReservedEvent
- Available: 65 (instant feedback)

If order fails:
- Release reservation
- Publish: StockReleasedEvent
- Available: 70 (restored)
```

**Key Learning:** Reservations are "soft locks". They hold inventory without permanent deduction. Allows rollback on failure.

---

### 3. **Event-Driven Inventory Synchronization**

```
Without events (synchronous):
Order Service calls Products Service (blocking):
POST /inventory/reserve {product_id, quantity}
└─ Must wait for response
└─ If timeout → retry → potential duplicate reservations
└─ Tight coupling: services know about each other

With events (asynchronous):
Order Service publishes: CartCheckoutInitiated
  ↓
Products Service subscribes, receives event
  ↓
Products Service reserves inventory, publishes: StockReservedEvent
  ↓
Order Service subscribes to StockReservedEvent
  ↓
Order Service creates order, publishes: OrderPlacedEvent
  ↓
Products Service confirms reservation

Benefits:
✅ Loose coupling: No direct service calls
✅ Resilient: Service can be down, events queue
✅ Idempotent: Duplicate events ignored via event_id
✅ Observable: Every step logged as event
```

**Key Learning:** Events decouple services. No service A → B → C chains. Instead: A publishes, B & C listen independently.

---

### 4. **Reservation Expiration & Cleanup**

```
Problem: Cart abandoned with reserved inventory
Solution: Reservations auto-expire after TTL

Workflow:
Customer adds 5 laptops to cart at 2:00 PM
  ↓
StockReservedEvent published
  ↓
Reservation created with expires_at = 2:00 PM + 24 hours
  ↓
Customer abandons cart, closes browser
  ↓
Next day, 2:00 PM: Cleanup job runs
  ↓
ExpireReservations() finds stale reservations
  ↓
Status: reserved → expired
  ↓
Stock available again

SQL:
UPDATE inventory_reservations
SET status = 'expired'
WHERE status = 'reserved' AND expires_at < NOW()
```

**Key Learning:** Distributed systems need auto-cleanup. TTLs prevent dead locks. Scheduled jobs or event-driven cleanup both work.

---

### 5. **Repository Pattern for Complex Queries**

```
Single Responsibility: Repository = Database Layer

ProductRepository methods:
- CreateProduct() → INSERT
- GetProduct(id) → SELECT WHERE id
- GetAllProducts() → SELECT with pagination
- UpdateProduct() → UPDATE
- DeleteProduct() → Soft DELETE
- DecrementStock() → UPDATE stock_quantity - quantity
- IncrementStock() → UPDATE stock_quantity + quantity

InventoryReservationRepository methods:
- CreateReservation() → INSERT
- GetReservation(id) → SELECT
- ReleaseReservation(id) → UPDATE status = released
- ExpireReservations() → UPDATE status = expired
- GetProductReservations(product_id) → SUM(quantity) WHERE status=reserved

Why separate?
- Each handles its domain
- Easy to test (mock repositories)
- Easy to swap implementations (SQL → MongoDB)
- Handler stays clean (no SQL)
```

**Key Learning:** Repository is the data access layer. Handler never touches database directly. Separation enables testability.

---

### 6. **Stock Validation Before Reservation**

```
Prevent overselling:

POST /inventory/reserve {product_id, quantity}

Step 1: Get product
  ├─ Find product by ID
  └─ Get stock_quantity = 100

Step 2: Calculate available
  ├─ Query active reservations
  ├─ SUM(quantity WHERE status = reserved) = 30
  └─ available = 100 - 30 = 70

Step 3: Check capacity
  ├─ Requested = 5
  ├─ available = 70
  └─ 70 ≥ 5? YES → Proceed
       OR 70 ≥ 5? NO → Return 409 Conflict

Step 4: Create reservation
  └─ Only if step 3 passed

Race condition prevention:
- Database constraint: Can't reserve more than exists
- Query includes FOR UPDATE lock (pessimistic locking)
- Or use optimistic locking with version field
```

**Key Learning:** Always validate before reserving. Prevent overselling at database level. Constraints > application logic.

---

### 7. **Event Routing Patterns**

```
RabbitMQ Topic Exchange: "products.events"

Routing keys:
- product.created     → Cart service: listen
- product.updated     → Cart/Orders: listen for price changes
- product.stock.reserved   → Cart: listen to confirm
- product.stock.released   → Order: listen to compensate

Binding:
products.events.queue --bind--> products.events (topic)
  Routing key pattern: "product.*" (match all product events)

Why topic instead of direct?
- product.created goes to: products.events.queue
- product.updated goes to: products.events.queue
- Multiple subscribers on same queue
- Each processes independently

If needed later:
- Add cart.events exchange for cart-specific events
- Add orders.events exchange for order-specific events
```

**Key Learning:** Topic exchanges with routing patterns = flexible event routing. One publisher, many subscribers.

---

### 8. **Idempotency at Service Level**

```
Without idempotency store:

Event: StockReservedEvent arrives
  ↓
Process: Reserve 5 units
  ↓
ACK sent to RabbitMQ
  ↓
Network blip: retry triggered
  ↓
SAME event arrives again
  ↓
Process: Reserve 5 MORE units (double reservation!)

With idempotency store (in Phase 4 setup):

Event: StockReservedEvent arrives (event_id: uuid-123)
  ↓
Check: SELECT * FROM idempotency_records WHERE event_id = uuid-123
  ↓
Not found → Process event
  ↓
INSERT INTO idempotency_records (event_id, service, action, result)
  ↓
ACK sent to RabbitMQ
  ↓
SAME event arrives again
  ↓
Check: SELECT * FROM idempotency_records WHERE event_id = uuid-123
  ↓
Found → SKIP processing
  ↓
ACK anyway (already processed)

Result: Single deduction, not double
```

**Key Learning:** Idempotency prevents duplicate processing. Every event must be traceable by event_id.

---

### 9. **Gin Handler Structure**

```
Standard handler pattern (Phase 4 uses):

func (ph *ProductHandler) ReserveInventory(c *gin.Context) {
    // 1. Parse request
    var req models.ReserveInventoryRequest
    if err := c.ShouldBindJSON(&req) {
        c.JSON(http.StatusBadRequest, errorResponse)
        return
    }
    
    // 2. Validate
    ctx := context.Background()
    product, err := ph.productRepo.GetProduct(ctx, req.ProductID)
    
    // 3. Business logic
    available := product.StockQuantity - reserved
    if available < req.Quantity {
        c.JSON(http.StatusConflict, errorResponse)
        return
    }
    
    // 4. Database operation
    reservation := models.NewInventoryReservation(...)
    ph.inventoryRepo.CreateReservation(ctx, reservation)
    
    // 5. Event publishing (async)
    event := events.StockReservedEvent{...}
    ph.eventPublisher.PublishProductEvent(ctx, event)
    
    // 6. Response
    c.JSON(http.StatusCreated, reservation)
}

Key pattern: Validate → Business Logic → DB → Events → Response
```

**Key Learning:** Clean handler flow = easy to understand and debug. Separation of concerns at every layer.

---

### 10. **Schema-Aware Queries**

```
All repositories use replaceSchema() helper:

query = `
    SELECT * FROM $schema.products WHERE id = $1
`
query = replaceSchema(query, "catalog")
// Result: SELECT * FROM catalog.products WHERE id = $1

Why needed?
- Same SQL, different schema per service
- Products uses: catalog schema
- Cart uses: cart schema
- Orders uses: orders schema
- DRY: Write once, use everywhere

Without this:
- Hardcode "SELECT * FROM catalog.products" in product service
- Hardcode "SELECT * FROM cart.carts" in cart service
- Hardcode "SELECT * FROM orders.orders" in orders service
- If schema name changes → update 10 places
```

**Key Learning:** Use template queries with schema placeholders. Enables multi-schema support without duplication.

---

### 11. **Error Response Standardization**

```
All endpoints return consistent error format:

Success (201):
{
  "message": "Inventory reserved successfully",
  "reservation": {...}
}

Client error (400):
{
  "error": "invalid request body",
  "message": "field validation error details",
  "code": 400
}

Validation error (400):
{
  "error": "insufficient stock",
  "message": "not enough inventory available",
  "code": 400
}

Not found (404):
{
  "error": "product not found",
  "message": "no product with id 999",
  "code": 404
}

Conflict (409):
{
  "error": "insufficient stock",
  "message": "requested 100, available 50",
  "code": 409
}

Server error (500):
{
  "error": "failed to reserve inventory",
  "message": "database error: connection timeout",
  "code": 500
}

Benefits:
✅ Frontend knows error structure
✅ Can parse and show friendly messages
✅ Logging/monitoring consistent
✅ API documentation clear
```

**Key Learning:** Standardized responses prevent frontend from guessing. Every service returns same error format.

---

### 12. **Asynchronous Event Publishing Pattern**

```
IMPORTANT: Events published AFTER database commit

Correct order:
1. Begin transaction
2. Update database (inventory reservation)
3. COMMIT transaction
4. Publish event to RabbitMQ

If error:
  - Transaction rolls back
  - Event never published
  - No inconsistency

Why this order?
- If event published BEFORE commit, then commit fails:
  - Other services received event
  - But database didn't change
  - Inconsistency!

- If event published AFTER commit:
  - Database changed first
  - Event is notification of already-committed fact
  - If RabbitMQ fails, event can retry
  - Other services catch up eventually

This is "transactional outbox pattern"
```

**Key Learning:** Events are notifications of completed actions. Publish after commit, not before.

---

## Architecture Overview After Phase 4

```
Products Service fully implements:

HTTP Request (Gin)
  ↓
Handler Layer (Parse, validate, respond)
  ↓
Repository Layer (Database queries)
  ├─ ProductRepository (CRUD products)
  ├─ CategoryRepository (CRUD categories)
  └─ InventoryReservationRepository (Track reservations)
  ↓
Database Layer (Schema-aware connections)
  ├─ Postgres: catalog schema
  └─ Idempotency store: track processed events
  ↓
Event Publishing (After DB commit)
  ├─ ProductCreated
  ├─ ProductUpdated
  ├─ StockReserved
  └─ StockReleased
  ↓
RabbitMQ Topic Exchange (products.events)
  ├─ Queue: products.events.queue
  └─ DLQ: products.events.dlq (poison messages)
```

---

## Critical Patterns for Later Phases

| Pattern | Where Used | Why Important |
|---------|-----------|---------------|
| **Event Publishing** | Cart, Orders, Users | Services communicate async |
| **Idempotency Tracking** | All services | Prevent duplicate processing |
| **Reservation Pattern** | Cart, Orders | Saga state management |
| **Repository Layer** | All services | Clean data access |
| **Schema Isolation** | All services | Service data independence |
| **Error Standardization** | All services | Consistent API contract |
| **Gin Handlers** | All services | HTTP layer pattern |

---

## What Phase 4 Taught You

✅ **Event publishing** as core async mechanism  
✅ **Inventory management** with reservations  
✅ **Stock tracking** across multiple services  
✅ **Event routing** with topic exchanges  
✅ **Repository pattern** for clean data layer  
✅ **Idempotency** importance in distributed systems  
✅ **Schema isolation** in multi-service databases  
✅ **Error handling** standardization  


