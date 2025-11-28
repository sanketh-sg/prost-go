## Phase 5: Cart Service - Key Learnings Summary

### 1. **Event Subscription Pattern (Consumer Side)**

```
Producer (Products Service) publishes:
StockReservedEvent → products.events exchange

Subscriber (Cart Service) receives:
products.events exchange 
  → (routing key: product.stock.*)
  → cart.events.queue (bindings allow multi-source)
  → Consumer reads message
  → Event handler processes
  → Idempotency check
  → Business logic
  → Record processed
  → ACK to RabbitMQ

Two-way communication example:
Products publishes: StockReservedEvent
  ↓
Cart listens, creates inventory lock
  ↓
Cart publishes: ItemAddedToCartEvent
  ↓
Orders listens, validates cart
  ↓
Orders publishes: OrderPlacedEvent
  ↓
Cart listens to confirm order

Result: Decoupled async workflow
```

**Key Learning:** Event subscription is the "listener" pattern. One service publishes, many listen. No coupling between them.

---

### 2. **Saga State Machine (Distributed Transaction Tracking)**

```
Checkout Saga States:

PENDING (initial)
  ├─ User clicks checkout
  ├─ CartCheckoutInitiated event published
  └─ Saga state created with correlation_id
  ↓
INVENTORY_VALIDATED (step 1)
  ├─ Cart service receives CartCheckoutInitiated
  ├─ Publishes CartCheckoutInitiated
  ├─ Orders service receives, reserves inventory
  ├─ Publishes StockReservedEvent
  └─ Saga status updated
  ↓
ORDER_CREATED (step 2)
  ├─ Orders service creates order
  ├─ Publishes OrderPlacedEvent
  ├─ Cart receives confirmation
  └─ Saga status updated
  ↓
CONFIRMED (success)
  ├─ All steps complete
  ├─ Order finalized
  └─ Saga complete

OR if failure at any step:
  ↓
FAILED
  ├─ Publishes OrderFailedEvent
  └─ Triggers compensations
  ↓
COMPENSATED
  ├─ StockReleasedEvent published
  ├─ Inventory unlocked
  ├─ Cart reverted
  └─ Saga complete (failed state)

Payload tracking:
{
  "order_id": 123,
  "cart_id": "uuid",
  "user_id": "user-uuid",
  "items": [...],
  "total": 5000.00,
  "step_1_timestamp": "2025-11-24T10:00:00Z",
  "step_2_timestamp": "2025-11-24T10:00:05Z"
}

Compensation log:
[
  "StockReserved id=res-1",
  "StockReserved id=res-2",
  "CartCleared"
]
```

**Key Learning:** Saga state is the "orchestration brain". It tracks what happened, where you are, and what to undo if things fail.

---

### 3. **Idempotency in Event Handling (Critical!)**

```
Without idempotency checking:

Event: StockReservedEvent (event_id: uuid-123)
  ↓
Cart subscriber processes:
  - Create inventory lock
  - Publish ItemAddedToCart
  ↓
Network delay, retry triggered
  ↓
SAME event arrives (same event_id: uuid-123)
  ↓
Cart subscriber processes AGAIN:
  - Create ANOTHER inventory lock (DUPLICATE!)
  - Publish ItemAddedToCart AGAIN

Result: Double inventory locked, double event published

WITH idempotency checking (Phase 5 implementation):

Event arrives (event_id: uuid-123)
  ↓
Check: SELECT * FROM idempotency_records 
       WHERE event_id = uuid-123 AND service_name = 'cart'
  ↓
Not found → Process event
  ↓
Business logic runs
  ↓
INSERT INTO idempotency_records 
  (event_id, service_name, action, result, timestamp)

Retry arrives (SAME event_id: uuid-123)
  ↓
Check: SELECT * FROM idempotency_records 
       WHERE event_id = uuid-123 AND service_name = 'cart'
  ↓
Found! → Skip processing
  ↓
Return success anyway (already done)

Result: Single inventory lock, event processed exactly once
```

**Key Learning:** Every event handler MUST check idempotency. Without it, retries cause corruption.

---

### 4. **Inventory Locks vs Reservations (Cart-Specific Pattern)**

```
Three levels of inventory tracking:

LEVEL 1: Product Service (catalog schema)
├─ Table: inventory_reservations
├─ Scope: Order-level (24hr TTL)
├─ Purpose: Track what orders reserved
├─ Example: Order #101 reserved 5 laptops
└─ Status: reserved → released → expired

LEVEL 2: Cart Service (cart schema)
├─ Table: inventory_locks
├─ Scope: Cart-level (1hr TTL)
├─ Purpose: Temporary holds while shopping
├─ Example: User's cart has 5 laptops "on hold"
└─ Status: locked → released → expired

LEVEL 3: Database Constraint
├─ Mechanism: Transaction isolation
├─ Scope: At commit time
├─ Purpose: Prevent overselling
└─ Example: SUM(reserved) + SUM(locked) ≤ stock

Timeline:
1. User adds laptop to cart
   → Create inventory_lock (1hr timeout)
   → Available = stock - locks - reservations
   
2. User checks out
   → Cart publishes CartCheckoutInitiated
   → Orders reserves from catalog
   → Create inventory_reservation (24hr timeout)
   → Orders publishes OrderPlaced

3. User abandons cart after 30min
   → Cart inventory_lock expires after 1hr
   → Automatically released
   → Available stock restored

4. Order ships
   → Inventory moves from "reserved" to "fulfilled"
   → Stock decremented in catalog.products
   → inventory_lock AND inventory_reservation deleted
```

**Key Learning:** Multiple levels of locks = safety. Carts use short TTL (1hr), orders use long TTL (24hr), products use atomic decrements.

---

### 5. **Correlation ID for Saga Tracing**

```
Single user action creates many events:

User clicks "Checkout" at 10:00:00
  ↓
Saga created with: correlation_id = "corr-uuid-abc-123"
  ↓
Event 1: CartCheckoutInitiated
  {
    event_id: "event-uuid-1",
    correlation_id: "corr-uuid-abc-123",  ← Same for all saga events
    timestamp: 10:00:00,
    data: {cart_id, user_id, total}
  }
  ↓
Event 2: StockReservedEvent
  {
    event_id: "event-uuid-2",
    correlation_id: "corr-uuid-abc-123",  ← Same!
    timestamp: 10:00:02,
    data: {product_id: 1, quantity: 5}
  }
  ↓
Event 3: StockReservedEvent
  {
    event_id: "event-uuid-3",
    correlation_id: "corr-uuid-abc-123",  ← Same!
    timestamp: 10:00:03,
    data: {product_id: 2, quantity: 3}
  }
  ↓
Event 4: OrderPlacedEvent
  {
    event_id: "event-uuid-4",
    correlation_id: "corr-uuid-abc-123",  ← Same!
    timestamp: 10:00:05,
    data: {order_id: 123}
  }

Query to trace entire saga:
SELECT * FROM event_log 
WHERE correlation_id = "corr-uuid-abc-123"
ORDER BY timestamp

Result: 4 events in sequence = complete user journey
Debugging becomes easy!
```

**Key Learning:** Correlation ID = saga fingerprint. Same ID across all events = traceable transaction.

---

### 6. **Asynchronous Event-Driven Checkout Flow**

```
SYNCHRONOUS (OLD WAY - blocking):
User clicks checkout
  ↓
Gateway calls Orders Service (wait)
  ├─ Orders calls Catalog Service (wait)
  │  └─ Check stock → reserve → return result
  ├─ Orders calls Cart Service (wait)
  │  └─ Validate items → return result
  └─ Return order confirmation
  ↓
Response time: 2-3 seconds (if all services fast)
Single failure = entire checkout fails

ASYNCHRONOUS (NEW WAY - event-driven):
User clicks checkout (instant response: "processing")
  ↓
CartCheckoutInitiated event published to RabbitMQ
  ↓
Cart service receives event
  → Publishes CartCheckoutInitiated
  → Logs saga state
  → Responds immediately

Orders service subscribes, processes asynchronously
  → Receives CartCheckoutInitiated
  → Validates cart
  → Reserves inventory from Products Service (REST call)
  → Publishes OrderPlacedEvent

Cart service receives OrderPlacedEvent
  → Updates saga status to CONFIRMED
  → Notifies user (via WebSocket or polling)

Products service receives requests
  → Creates inventory_reservation
  → Publishes StockReservedEvent
  → Other services listen

Timeline:
10:00:00.000 - User clicks checkout
10:00:00.050 - Saga created, event published, response sent
10:00:00.100 - Orders service processes (async)
10:00:00.200 - Products service processes (async)
10:00:00.500 - Saga complete, user notified

Benefits:
✅ Response time: 50ms (vs 2-3s)
✅ Scale: If Orders slow, doesn't block checkout
✅ Resilience: Service down? Events queue up
✅ Observability: Every step is an event
```

**Key Learning:** Async events scale better. Response ≠ Processing. User gets quick feedback, processing happens in background.

---

### 7. **Saga Compensation Log (Rollback Tracking)**

```
Problem: Order fails mid-saga. What was already done?

Solution: Track every non-idempotent action

Compensation log example:

Step 1: Reserve Product 1 (5 laptops)
  ├─ Action: StockReservedEvent published
  ├─ Compensation: StockReleasedEvent
  └─ Log: ["StockReserved:laptop:5"]

Step 2: Reserve Product 2 (3 monitors)
  ├─ Action: StockReservedEvent published
  ├─ Compensation: StockReleasedEvent
  └─ Log: ["StockReserved:laptop:5", "StockReserved:monitor:3"]

Step 3: Create order record
  ├─ Action: INSERT INTO orders
  ├─ Compensation: None (just mark failed)
  └─ Log: unchanged (idempotent if retry)

Step 4: Process payment (FAILS)
  ├─ Action: Failed
  ├─ Compensation: Release everything
  └─ Rollback sequence:
      - Release monitor reservation (step 2)
      - Release laptop reservation (step 1)
      - Mark order as cancelled
      - Publish OrderCancelledEvent

SQL to rollback:
UPDATE orders SET status = 'cancelled' WHERE order_id = 123;
DELETE FROM inventory_reservations WHERE order_id = 123;
PUBLISH OrderCancelledEvent;

Database provides atomicity for one service.
Compensation log provides atomicity across services.
```

**Key Learning:** Compensation log = rollback journal. Know what to undo when things fail.

---

### 8. **Repository Pattern with Multiple Repositories**

```
Cart Service has 3 repositories:

CartRepository
├─ GetCart(cartID) → Fetch cart + items
├─ AddItem(cartID, productID, qty, price)
├─ RemoveItem(cartID, productID)
├─ UpdateCartTotal(cartID, total)
└─ ClearCart(cartID)

SagaStateRepository
├─ CreateSagaState(correlationID, sagaType)
├─ GetSagaState(correlationID)
├─ UpdateSagaStatus(correlationID, status)
├─ AddCompensation(correlationID, compensation)
└─ UpdateSagaPayload(correlationID, payload)

InventoryLockRepository
├─ CreateLock(cartID, productID, qty, reservationID)
├─ GetLocksByCartID(cartID)
├─ ReleaseLock(reservationID)
├─ ReleaseCartLocks(cartID)
└─ ExpireLocks() → Cleanup old locks

Why multiple?
- Separation of concerns
- Each handles one entity
- Easy to understand
- Easy to test
- Easy to parallelize

Handler coordinates:
CheckoutCart(cartID)
  ├─ cartRepo.GetCart() ← fetch cart
  ├─ sagaRepo.CreateSagaState() ← create saga
  ├─ cartRepo.UpdateCartStatus() ← mark checked out
  ├─ eventPublisher.PublishCartEvent() ← send event
  └─ Return 202 Accepted

Later, when event arrives:
EventHandler.HandleStockReserved()
  ├─ inventoryLockRepo.CreateLock() ← lock inventory
  ├─ sagaRepo.UpdateSagaStatus() ← update saga
  ├─ idempotencyStore.RecordProcessed() ← mark idempotent
  └─ Return success
```

**Key Learning:** Multiple repositories = clean code. Each owns its domain. Handler orchestrates them.

---

### 9. **Message Subscription Handler Function**

```
RabbitMQ Subscriber pattern:

subscriber := messaging.NewSubscriber(rmqConn, "cart.events.queue")

// Pass handler function
subscriber.Subscribe(func(message []byte) error {
    return eventHandler.HandleEvent(ctx, message)
})

Handler function signature:
func(message []byte) error

Flow:
1. Message arrives from queue
2. Handler function called with raw bytes
3. Handler returns:
   - nil ← Message ACK'd (success)
   - error ← Message NACK'd (retry/DLQ)

Implementation:
func (eh *EventHandler) HandleEvent(ctx context.Context, message []byte) error {
    // 1. Parse event
    var event BaseEvent
    if err := json.Unmarshal(message, &event) {
        return err  // NACK
    }
    
    // 2. Check idempotency
    processed, _ := eh.idempotencyStore.IsProcessed(ctx, event.EventID, "cart")
    if processed {
        return nil  // ACK anyway (already done)
    }
    
    // 3. Route by type
    switch event.EventType {
        case "StockReserved":
            return eh.handleStockReserved(message)
        case "OrderPlaced":
            return eh.handleOrderPlaced(message)
        default:
            return nil  // Ignore unknown types
    }
}

Error handling:
- Return nil = ACK = message discarded (success)
- Return error = NACK = message goes to DLQ or retries
- Timeout (no return) = redelivery after timeout

Best practice:
- Always check idempotency first
- Log before processing
- Return nil even if "already processed"
- Return error only on real failures
```

**Key Learning:** Handler function is the event entry point. Idempotency check first, then route to specific handler.

---

### 10. **Cart Total Calculation Pattern**

```
Whenever item added/removed, recalculate total:

AddItem flow:
1. Add item to cart_items table
2. Get cart and all items
3. Calculate: SUM(price * quantity) for all items
4. UPDATE carts.total = calculated_sum
5. Publish ItemAddedToCart event

Example:
Initial cart total: $0

User adds Laptop ($999.99 × 1)
  → SUM = 999.99
  → UPDATE carts SET total = 999.99
  
User adds Monitor ($299.99 × 2)
  → SUM = 999.99 + (299.99 × 2) = 1599.97
  → UPDATE carts SET total = 1599.97

User removes Laptop
  → SUM = 299.99 × 2 = 599.98
  → UPDATE carts SET total = 599.98

Why track at cart level?
- Fast lookup (no JOIN needed)
- No recalculation at checkout
- Snapshot prices (immutable in cart_items)
- If product price changes, cart price unchanged

Price snapshot important:
Product table: price = $999.99 (can change)
Cart item: price = $899.99 (frozen at add time)

If product discounted:
- New products get new price
- Existing carts keep original price
- When checkout → use cart prices, not current prices
- Fair to customer (got price when they added)

Query pattern:
SELECT 
  SUM(price * quantity) as total
FROM cart_items
WHERE cart_id = $1;

Database calculates, not application.
Prevents rounding errors in app code.
```

**Key Learning:** Always store total in cart. Recalculate on every add/remove. Use price snapshots for fairness.

---

### 11. **Event Publishing After Database Commit**

```
WRONG order (causes inconsistency):

1. Publish ItemAddedToCart event
2. Commit to database

If DB fails after publishing:
- Other services received event
- But database never saved the item
- Other services think item was added, but it wasn't

CORRECT order (Phase 5 uses this):

1. BEGIN TRANSACTION
2. INSERT INTO cart_items (...)
3. UPDATE carts SET total = ... 
4. COMMIT TRANSACTION
5. Publish ItemAddedToCart event

If DB fails at step 4:
- Event never published
- Other services don't know
- Retry will try again (idempotent)

If DB succeeds but event publishing fails:
- Cart item was saved
- Event will retry (RabbitMQ retry)
- Eventually publishes
- Acceptable (eventual consistency)

This is "Transactional Outbox" pattern:
Events are notifications of completed DB changes.
Not triggers that cause DB changes.
```

**Key Learning:** Events are notifications, not triggers. Publish after commit, not before.

---

## Architecture After Phase 5

```
Complete event-driven checkout:

User Interface
  ↓ (click checkout)
Cart Service
  ├─ HTTP Handler (CheckoutCart)
  ├─ Create Saga State (correlation_id)
  ├─ Publish CartCheckoutInitiated
  └─ Publish ItemAddedToCart
  
RabbitMQ (cart.events exchange)
  ├─ Queue: cart.events.queue
  ├─ Queue: orders.events.queue (subscribed to cart.*)
  └─ Queue: products.events.queue

Orders Service (subscribes to cart.*)
  ├─ Event Subscriber
  ├─ Receives CartCheckoutInitiated
  ├─ Validate cart (sync REST call to Cart)
  ├─ Reserve inventory (sync REST call to Products)
  ├─ Create order
  ├─ Publish OrderPlacedEvent
  └─ Publish OrderConfirmedEvent

Products Service (publishes to products.events)
  ├─ Receive inventory reservation request
  ├─ Check stock availability
  ├─ Create inventory_reservation
  ├─ Publish StockReservedEvent
  └─ (or StockReleasedEvent if fails)

Cart Service (subscribes to products.*)
  ├─ Event Subscriber
  ├─ Receives StockReservedEvent
  ├─ Create inventory_lock
  ├─ Update saga state
  └─ Handles StockReleasedEvent (cleanup)

Database Layer:
├─ cart schema (carts, cart_items, inventory_locks, saga_states)
├─ catalog schema (products, inventory_reservations)
├─ orders schema (orders, order_items) - Phase 6
└─ users schema (users) - Phase 3
```

---

## Critical Patterns Learned

| Pattern | Purpose | Example |
|---------|---------|---------|
| **Event Subscription** | Listen to other services | Cart listens to Products stock events |
| **Saga State Machine** | Track multi-step transactions | Checkout saga: pending → confirmed |
| **Idempotency Checking** | Prevent duplicate processing | Check event_id before processing |
| **Correlation ID** | Link related events | All checkout events share same correlation_id |
| **Compensation Log** | Track rollbacks | StockReserved → remember to release |
| **Inventory Locks** | Temporary holds | Cart uses 1hr TTL, Orders uses 24hr |
| **Event Publishing After Commit** | Consistency | DB first, then notify others |
| **Multiple Repositories** | Clean separation | CartRepo, SagaRepo, LockRepo |
| **Price Snapshots** | Fairness | Cart saves price at add time |
| **Event Type Routing** | Handler orchestration | Switch on event_type, call handler |

---

## What Phase 5 Taught You

✅ **Event subscription** as consumer of async flows  
✅ **Saga state management** for distributed transactions  
✅ **Idempotency checking** at handler entry point  
✅ **Inventory locks** for temporary reservations  
✅ **Correlation IDs** for saga tracing  
✅ **Compensation logging** for rollbacks  
✅ **Repository coordination** in event handlers  
✅ **Eventual consistency** in distributed systems  

---