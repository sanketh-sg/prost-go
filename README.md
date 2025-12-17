# Core Architecture

## Microservices:
User Service: Authentication and user management (JWT-based)
Products Service: Product catalog, categories, and inventory management
Cart Service: Shopping cart operations and item management
Orders Service: Order creation and saga-based distributed transactions

## Communication & Data:
GraphQL Gateway: Aggregates all microservices into a single API endpoint
RabbitMQ: Asynchronous event-driven communication between services (stock reserved, cart checkout initiated, orders confirmed)
PostgreSQL: Database with 4 separate schemas (one per service) following database-per-service pattern
Redis: Optional caching layer

## Key Architectural Patterns
Saga Pattern: Distributed transactions across Cart → Orders → Products services with compensation logic for failures
Event-Driven Architecture: Services communicate via RabbitMQ topic exchanges with routing keys (product.*, cart.*, order.*)
Idempotency: Prevents duplicate event processing through idempotency keys and records
Database-Per-Service: Each microservice owns its schema, ensuring loose coupling


Plan: Step-by-Step Implementation Roadmap
Current State: Gateway deleted, 4 empty service directories, infrastructure ready (PostgreSQL, Redis, RabbitMQ), frontend Vue scaffolded.

Phase 1: Foundation (Days 1-2)
1.1 Initialize Go modules & shared packages

Create go.mod at root, shared/go.mod for shared code
Implement models with DTOs (Product, Cart, Order, User, Event)
Implement events with event type definitions (event_id, timestamp, version fields)

1.2 Setup database connectivity layer
Create db with PostgreSQL connection manager (schema-aware)
Add connection pooling, prepared statements, migration runner
Support separate schemas: catalog, cart, orders, users


1.3 Setup RabbitMQ messaging layer
Create messaging with publisher/subscriber patterns
Implement DLQ (dead-letter queue) setup for each queue
Add idempotency tracking via event_id


Phase 2: Database Schema (Days 2-3)


2.1 Create migration files

001_create_users_schema.sql — Users table, idempotency table
002_create_catalog_schema.sql — Products, categories, inventory tracking
003_create_cart_schema.sql — Cart items, saga state
004_create_orders_schema.sql — Orders, order items, saga state
2.2 Setup RabbitMQ topology

Create exchanges: products.events, orders.events, cart.events
Create queues per service with DLQ bindings
Document queue naming conventions
2.3 Update docker-compose.yml

Add users service (port 8083)
Update DATABASE_URL per service with schema names
Add init script to create RabbitMQ exchanges/queues on startup
Phase 3: Users Service (Days 3-4)
3.1 Scaffold users service

Create services/users/main.go, services/users/go.mod
REST endpoints: POST /register, POST /login, GET /profile/:id
JWT token generation & validation middleware
3.2 Implement user database layer

User CRUD operations with password hashing (bcrypt)
Idempotency tracking for duplicate requests
3.3 Add to docker-compose

Build & test locally with docker-compose up users
Phase 4: Products Service (Days 4-6)
4.1 Scaffold products service

Create services/products/main.go, services/products/go.mod
REST endpoints: POST /products, GET /products, GET /products/:id, PATCH /products/:id
4.2 Implement product database layer

Product CRUD, inventory management
Idempotency tracking table
4.3 Implement event publishing

Publish ProductCreated, ProductUpdated, StockReserved, StockReleased events
Ensure event_id uniqueness for idempotency
4.4 Test event flow

Verify events published to RabbitMQ queue
Check DLQ behavior with malformed messages
Phase 5: Cart Service (Days 6-8)
5.1 Scaffold cart service

Create services/cart/main.go, services/cart/go.mod
REST endpoints: POST /cart, POST /cart/items, GET /cart/:id, DELETE /cart/:id
5.2 Implement cart database layer

Cart items, saga state tracking table
Idempotency tracking
5.3 Implement event consumers

Subscribe to StockReserved from products service
Subscribe to StockReleased for compensation
Update inventory locks in local DB
5.4 Implement event publishers

Publish ItemAddedToCart, CartCheckoutInitiated, CartCleared events
5.5 Integration testing

Test product stock → cart item flow with events
Phase 6: Orders Service & Saga Pattern (Days 8-11)
6.1 Scaffold orders service

Create services/orders/main.go, services/orders/go.mod
REST endpoint: POST /orders (saga initiator)
Saga state machine: PENDING → CART_VALIDATED → PAYMENT_PROCESSED → CONFIRMED (or FAILED → COMPENSATED)
6.2 Implement order database layer

Orders table, order items, saga state tracking
Idempotency tracking with compensation log
6.3 Implement saga orchestrator

Listen to CartCheckoutInitiated event
Step 1: Validate cart & reserve inventory from products service
Step 2: Create order, publish OrderPlaced event
Step 3: Handle compensations (OrderCancelled → release inventory, OrderRollback)
6.4 Implement event publishers & consumers

Publish: OrderPlaced, OrderConfirmed, OrderFailed, OrderCancelled
Consume: CartCheckoutInitiated, PaymentProcessed (future payment service)
6.5 Test saga flow

Happy path: cart → order created → inventory reserved
Sad path: order fails → inventory released (compensation)
DLQ handling: poison messages processed correctly
Phase 7: API Gateway (Days 11-13)
7.1 Scaffold GraphQL gateway

Create gateway/go.mod, gateway/main.go
Setup Apollo/GraphQL server in Go (gqlgen or graphql-go)
7.2 Implement GraphQL schemas

Queries: products, product(id), cart(id), orders, order(id), user(id)
Mutations: createProduct, addToCart, checkout (saga trigger), registerUser, loginUser
7.3 Implement resolvers

Route queries to respective microservices
Users service: synchronous REST calls
Products/Cart/Orders: HTTP to service APIs
Handle async saga callbacks (WebSocket subscriptions for order status)
7.4 Add authentication middleware

JWT validation from users service
Authorization checks per resolver
Phase 8: Frontend Integration (Days 13-14)
8.1 Create API client

Vue composables for GraphQL queries/mutations
Handle authentication (login → store JWT)
8.2 Implement UI pages

Product list, product detail, cart, checkout, orders, user profile
8.3 End-to-end testing

Full user journey: register → browse products → add to cart → checkout → order confirmation
Quick Reference: Which Service Handles What
Component	Service	Pattern
User authentication	Users	Synchronous REST + JWT
Product catalog	Products	Async events (stock changes)
Shopping cart	Cart	Async events + saga state
Order creation	Orders	Saga orchestrator (choreography)
Event publishing	All services	RabbitMQ with DLQ
API aggregation	Gateway	GraphQL + REST to services
Start with: Phase 1 (shared packages) → Phase 2 (database schema) → Phase 3 (users service). All prerequisite for later phases.