
## Order successful path

```
User adds items to cart
    ↓
POST /carts/:id/checkout
    ↓
Cart Service publishes CartCheckoutInitiatedEvent
    ↓
Orders Service receives event → Creates order
    ↓
Orders Service publishes OrderPlacedEvent
    ↓
Products Service receives event → Reserves inventory
    ↓
Products Service publishes StockReservedEvent
    ↓
Cart Service RECEIVES StockReservedEvent
    ├─ handleStockReserved() called
    ├─ Creates InventoryLock record
    └─ Updates saga status → "inventory_locked"
    
    Orders Service RECEIVES StockReservedEvent
    ├─ Updates saga status → "inventory_reserved"
    └─ Order complete ✓

Result: Order confirmed, inventory locked, cart updated
```

## Failed order path

```
Order processing fails (e.g., payment rejected)
    ↓
Orders Service publishes OrderFailedEvent
    ↓
Cart Service RECEIVES OrderFailedEvent
    ├─ handleOrderFailed() called
    └─ Updates saga status → "compensation_in_progress"
    
Orders Service publishes StockReleaseEvent
    ↓
Products Service releases inventory
    ↓
Cart Service RECEIVES StockReleasedEvent
    ├─ handleStockReleased() called
    ├─ Removes InventoryLock record
    └─ Updates saga status → "failed"

Result: Order cancelled, inventory released, cart cleaned up
```

Cart Service (EventHandler)
├─ RECEIVES events from other services
├─ UPDATES ITS OWN STATE (inventory locks, saga status)
├─ STORES information (doesn't make saga decisions)
└─ RESPONDS to orchestrator's instructions

The service that owns the primary resource (Orders owns orders) should orchestrate the saga involving that resource.

│   ├── Schema: cart
│   │   ├── carts   id | user_id | status | total | created_at | updated_at | abandoned_at 
│   │              ----+---------+--------+-------+------------+------------+--------------
│   │   └── cart_items   id | cart_id | product_id | quantity | price | created_at | updated_at 
│   │                   ----+---------+------------+----------+-------+------------+------------
│   │   └── idempotency_records  id | event_id | service_name | action | result | created_at 
│   │                           ----+----------+--------------+--------+--------+------------
│   │   └── saga_states   id | correlation_id | saga_type | status | order_id | payload | compensation_log | created_at | updated_at | expires_at 
│   │                    ----+----------------+-----------+--------+---------+---------+------------------+------------+------------+------------
│   │   └── inventory_locks   id | cart_id | product_id | quantity | reservation_id | status | locked_at | expires_at | released_at 
│   │                        ----+---------+------------+----------+----------------+--------+-----------+------------+-------------