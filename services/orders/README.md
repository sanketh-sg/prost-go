Use parameterized queries to prevent sql injection, The database driver handles escaping automatically, preventing SQL injection attacks.

```
1. Create Order object in Go
   order := &Order{
       ID: 123,
       UserID: "user-456",
       CartID: "cart-789",
       Total: 599.99,
       Status: "pending",
       CreatedAt: now,
       UpdatedAt: now,
   }

2. Call CreateOrder(ctx, order)
   ↓
3. Build SQL with placeholders ($1, $2, ...)
   ↓
4. Execute INSERT ... RETURNING
   ↓
5. Database inserts row and returns it
   ↓
6. Scan() maps returned columns back to order struct
   ↓
7. Return order with confirmed values from database
```

When a user checks out (Cart → Orders → Products), you need to coordinate actions across multiple services. If something fails midway, you need to rollback previous steps. This is a distributed transaction problem.

How SagaStateRepository helps is, it tracks the entire saga lifecycle by storing saga state in the database.

Method	            Purpose

CreateSagaState()	Initialize saga when checkout starts (status: PENDING)
GetSagaState()	    Retrieve current saga state by correlation ID
UpdateSagaStatus()	Progress saga state (PENDING → CONFIRMING → CONFIRMED)
UpdateSagaOrderID()	Link saga to actual order once created
AddCompensation()	Log compensation actions (e.g., "release_inventory_123")
UpdateSagaPayload()	Store intermediate results from each service

## Complete Order Service Workflow

```
User checks out (Cart Service)
    ↓
[CartCheckoutInitiatedEvent] → RabbitMQ
    ↓
Orders Service receives event
    ↓
1. Create Order (status: pending)
2. Update Saga State → cart_validated
3. Publish OrderPlacedEvent
    ↓
Products Service receives OrderPlacedEvent
    ↓
4. Reserve Inventory → [StockReservedEvent]
    ↓
Orders Service receives StockReservedEvent
    ↓
5. Record Inventory Reservation
6. Update Saga State → inventory_reserved
    ↓
Order Complete 

```

Orders Service (SagaOrchestrator)
├─ INITIATES the saga (creates order)
├─ PUBLISHES events to trigger other services
├─ TRACKS saga state (pending → confirmed → failed)
├─ MAKES DECISIONS (what to do next)
└─ LOGS COMPENSATION (what to undo if failure)