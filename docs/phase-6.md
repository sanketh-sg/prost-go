Phase 6: Key Learnings — Saga Orchestration
1. Event Routing is the Core Pattern
The saga orchestrator routes incoming events by type (CartCheckoutInitiated, StockReserved, StockReleased) and dispatches to specific handlers. This decouples event processing logic from event consumption—each handler focuses on one step of the distributed transaction.

2. Saga State is the Transaction Memory
The SagaState record stores correlation IDs, user/cart info, and status progression (pending → cart_validated → inventory_reserved → confirmed). Without this, the saga can't know "where am I in the transaction?" when the next event arrives.

3. Compensation Logging Enables Rollback
When inventory is reserved, we immediately log a compensation action: "if saga fails, release this reservation." This creates a rollback plan BEFORE anything goes wrong—critical for atomicity across services.

4. Idempotency Prevents Double-Processing
Even though we subscribe to the same event queue multiple times (in different services), the idempotency check (IsProcessed()) ensures each event is processed exactly once per service. Without this, a cart checkout could create 2 orders.

5. OrderID is Created Fresh, Not from Cart
The order gets its own ID (UUID-based) rather than reusing cart ID. This decouples order lifecycle from cart—orders can exist independently and be retrieved/cancelled without affecting cart state.

6. Failure Publishing Triggers Compensation
When order creation fails, we publish OrderFailedEvent immediately. This signals other services to undo their work—the saga recognizes the failure and compensation begins automatically (without waiting for a timeout).

7. Saga Orchestration (not Choreography)
Orders Service is the "conductor"—it makes decisions about what happens next based on events. Cart and Products services don't know about each other; they just react to orders. This centralized control makes saga state easier to debug and understand.

8. Multi-Repository Coordination
One saga handler uses 4 repositories (order, saga, compensation, inventory). This reveals that business logic often spans multiple domain models—the orchestrator is the glue that coordinates them as a single distributed transaction.