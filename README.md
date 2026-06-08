# Prost-Go: Microservices E-Commerce Platform

A production-grade microservices e-commerce platform built with Go, Vue.js, and GraphQL. This project demonstrates modern distributed systems patterns including saga-based transactions, event-driven architecture, and database-per-service design.

---

## 📊 Project Status

**Current Phase:** Phase 1 - Foundation (In Progress)

| Phase | Title | Status | Timeline |
|-------|-------|--------|----------|
| 1 | Foundation (Shared packages, DB layer, Messaging setup) | 🔄 In Progress | Days 1-2 |
| 2 | Database Schemas & RabbitMQ Topology | ⏳ Planned | Days 2-3 |
| 3 | Users Service (Authentication & JWT) | ⏳ Planned | Days 3-4 |
| 4 | Products Service (Catalog & Inventory) | ⏳ Planned | Days 4-6 |
| 5 | Cart Service (Shopping cart with events) | ⏳ Planned | Days 6-8 |
| 6 | Orders Service & Saga Pattern | ⏳ Planned | Days 8-11 |
| 7 | GraphQL API Gateway | ⏳ Planned | Days 11-13 |
| 8 | Frontend Integration | ⏳ Planned | Days 13-14 |

**Current State:**
- ✅ Core architecture designed
- ✅ Infrastructure ready (PostgreSQL, Redis, RabbitMQ in docker-compose)
- ✅ Frontend scaffolded (Vue)
- 🔄 Service directories created (empty)
- 🔄 Implementing shared packages and database connectivity

---

## 🏗️ Architecture Overview

### Microservices

| Service | Purpose | Port | Pattern |
|---------|---------|------|---------|
| **Users** | Authentication, user management (JWT-based) | 8083 | Synchronous REST |
| **Products** | Product catalog, categories, inventory | 8081 | Async events |
| **Cart** | Shopping cart operations, item management | 8082 | Async events + Saga |
| **Orders** | Order creation, saga orchestration | 8084 | Saga Orchestrator |
| **Gateway** | GraphQL API aggregation layer | 8080 | GraphQL |

### Communication & Data Layer

```
┌─────────────────────────────────────┐
│      GraphQL Gateway (Port 8080)    │
└────────┬────────────────────────────┘
         │
    ┌────┴────┬────────┬──────────┐
    │          │        │          │
  REST       REST      REST       REST
    │          │        │          │
┌──┴─┐  ┌────┴──┐  ┌──┴─┐  ┌────┴──┐
│Users│  │Product│  │Cart│  │Orders │
└──┬──┘  └───┬───┘  └──┬─┘  └───┬───┘
   │         │         │        │
   └─────────┴────┬────┴────────┘
              ┌──┴────────────┐
              │   RabbitMQ    │
              │  Event Bus    │
              └───────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
    ┌───┴─────────┬────────┬─────┴───┐
    │             │        │         │
PostgreSQL    Redis      Cache    Outbox
(4 schemas)             (optional) Tables
```

### Key Technologies

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend** | Go 1.21+ | Service implementations |
| **API** | GraphQL (gqlgen) | Unified API layer |
| **Message Queue** | RabbitMQ | Event-driven communication |
| **Database** | PostgreSQL | Persistent data (4 schemas) |
| **Cache** | Redis | Optional performance optimization |
| **Frontend** | Vue 3 | Web UI |

---

## 🎯 Architecture Patterns

### 1. **Saga Pattern** (Distributed Transactions)
Manages multi-service transactions with compensation logic:
- **Happy Path:** Cart → Reserve Stock → Create Order → Confirm
- **Failure Path:** Automatic compensation (release stock, rollback)
- **Implementation:** Orchestration via Orders service

### 2. **Event-Driven Architecture**
Asynchronous communication via RabbitMQ:
- Topic exchanges: `products.*`, `cart.*`, `orders.*`
- Routing keys for selective message delivery
- Dead-Letter Queues (DLQ) for failed messages
- Idempotency keys prevent duplicate processing

### 3. **Database-Per-Service**
Each microservice owns its database schema:
- **users** schema: User accounts, authentication
- **catalog** schema: Products, categories, inventory
- **cart** schema: Cart items, saga state
- **orders** schema: Orders, order items, saga state
- Benefits: Loose coupling, independent scaling, schema evolution

### 4. **API Gateway Pattern**
GraphQL Gateway aggregates all services:
- Single entry point for clients
- Service discovery abstraction
- Authentication enforcement
- Request routing and composition

### 5. **Outbox Pattern** (Dual Write Solution)
Prevents message loss on database write + publish failures:
- Write entity AND event to database (same transaction)
- Polling service publishes events from outbox table
- Retry mechanism ensures eventual consistency
- Trade-off: ~6s latency (polling interval)

---

## 🔧 Technology Stack

```
Language Composition:
- Go:         84.9%  (Core services)
- Vue:        10.0%  (Frontend UI)
- TypeScript:  3.9%  (Frontend typing)
- Other:       1.2%  (Config, docs)
```

### Dependencies
- **graphql-go** or **gqlgen** - GraphQL server
- **amqp** - RabbitMQ client
- **pgx** - PostgreSQL driver
- **redis** - Redis client
- **jwt-go** - JWT token handling
- **bcrypt** - Password hashing

---

## 📋 Implementation Roadmap

### Phase 1: Foundation (Days 1-2) - **CURRENT**

**1.1 Go Modules & Shared Packages**
- [ ] Initialize `go.mod` at root
- [ ] Create `shared/go.mod` for shared code
- [ ] Implement data models (Product, Cart, Order, User, Event)
- [ ] Define event types with versioning

**1.2 Database Connectivity Layer**
- [ ] Create PostgreSQL connection manager
- [ ] Implement connection pooling
- [ ] Add migration runner
- [ ] Support schema-aware operations

**1.3 RabbitMQ Messaging Layer**
- [ ] Implement publisher/subscriber patterns
- [ ] Setup Dead-Letter Queues (DLQ)
- [ ] Add idempotency tracking via event_id

---

### Phase 2: Database Schemas (Days 2-3)

**2.1 Migration Files**
```
001_create_users_schema.sql
  - users table
  - idempotency_records table
  
002_create_catalog_schema.sql
  - products table
  - categories table
  - inventory_tracking table
  
003_create_cart_schema.sql
  - cart_items table
  - saga_state table
  
004_create_orders_schema.sql
  - orders table
  - order_items table
  - saga_state table
```

**2.2 RabbitMQ Topology**
- [ ] Create topic exchanges (products.events, orders.events, cart.events)
- [ ] Create queues per service with DLQ bindings
- [ ] Document queue naming conventions

**2.3 Docker Compose Updates**
- [ ] Add users service (port 8083)
- [ ] Configure DATABASE_URL per service
- [ ] Add RabbitMQ init script

---

### Phase 3: Users Service (Days 3-4)
- [ ] REST endpoints: POST /register, POST /login, GET /profile/:id
- [ ] JWT token generation & validation
- [ ] Password hashing (bcrypt)
- [ ] Docker integration

### Phase 4: Products Service (Days 4-6)
- [ ] REST endpoints: CRUD operations
- [ ] Inventory management
- [ ] Event publishing (ProductCreated, StockReserved, StockReleased)
- [ ] Idempotency tracking

### Phase 5: Cart Service (Days 6-8)
- [ ] REST endpoints: Cart operations
- [ ] Event consumers (StockReserved, StockReleased)
- [ ] Saga state tracking
- [ ] Integration testing

### Phase 6: Orders Service & Saga (Days 8-11)
- [ ] Saga orchestrator
- [ ] Event publishing/consuming
- [ ] Compensation logic
- [ ] End-to-end saga testing

### Phase 7: GraphQL Gateway (Days 11-13)
- [ ] GraphQL schema definition
- [ ] Resolver implementation
- [ ] Service routing
- [ ] Authentication middleware
- [ ] WebSocket subscriptions (order status updates)

### Phase 8: Frontend Integration (Days 13-14)
- [ ] API client (Vue composables)
- [ ] UI pages (products, cart, checkout, orders, profile)
- [ ] End-to-end user journey
- [ ] Authentication flow

---

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+ (for frontend)

### Setup

```bash
# Clone repository
git clone https://github.com/sanketh-sg/prost-go.git
cd prost-go

# Start infrastructure (PostgreSQL, RabbitMQ, Redis)
docker-compose up -d

# Initialize database
# (Migrations runner to be implemented in Phase 2)

# Run services (after Phase 1 implementation)
go run ./services/users
go run ./services/products
go run ./services/cart
go run ./services/orders
go run ./gateway

# Start frontend
cd frontend
npm install
npm run dev
```

### Testing GraphQL
Visit Apollo Studio: https://studio.apollographql.com/dev
- GraphQL URL: `http://localhost:8080/graphql`
- Or use curl:
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ products { id name } }"}'
```

---

## 📚 Quick Reference: Service Responsibilities

| Component | Service | Pattern | Communication |
|-----------|---------|---------|----------------|
| User authentication | Users | Sync REST + JWT | HTTP |
| Product catalog | Products | Async events | RabbitMQ |
| Shopping cart | Cart | Async + Saga | RabbitMQ + HTTP |
| Order creation | Orders | Saga orchestrator | RabbitMQ + HTTP |
| Event publishing | All services | Async | RabbitMQ |
| API aggregation | Gateway | GraphQL | GraphQL + HTTP |

---

## ⚙️ Important Design Decisions

### Dual Write Problem Solution
**Challenge:** Writing to database AND publishing events atomically

**Solutions Evaluated:**
1. **Exponential Backoff (CURRENT)** ✅
   - After DB write, spawn goroutine with retry logic
   - RabbitMQ persistence ensures durability
   - Pros: Simple, low latency (~100ms)
   - Cons: Requires downstream idempotency

2. **Outbox Pattern**
   - Write entity and event to DB in same transaction
   - Polling service publishes from outbox table
   - Pros: Guaranteed atomicity
   - Cons: Higher latency (~6s), operational complexity

3. **Change Data Capture (CDC)**
   - Database triggers generate events automatically
   - Pros: Elegant, automatic
   - Cons: Database-specific, complex setup

**Downstream Requirement:** All services MUST handle duplicate events via idempotency keys

---

## 📖 Documentation Structure

- `README.md` - This file (overview & roadmap)
- `gateway/README.md` - GraphQL gateway implementation details
- `frontend/README.md` - Vue frontend development guide
- Service READMEs (to be created in Phases 3-6)

---

## 🤝 Contributing

Phases are strictly sequential. Start with Phase 1 foundations before moving to later phases.

Development order:
1. Phase 1: Shared packages → DB layer → Messaging
2. Phase 2: Database schemas → RabbitMQ topology
3. Phase 3: Users service (simplest, no events)
4. Phase 4: Products service (event publishing)
5. Phase 5: Cart service (event consuming)
6. Phase 6: Orders service (saga orchestration)
7. Phase 7: Gateway (API aggregation)
8. Phase 8: Frontend (UI integration)

---

## 📝 License

MIT

---

## 🔗 Resources

- [Saga Pattern](https://microservices.io/patterns/data/saga.html)
- [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html)
- [Database-per-Service](https://microservices.io/patterns/data/database-per-service.html)
- [Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [RabbitMQ Best Practices](https://www.rabbitmq.com/best-practices.html)
- [GraphQL Go](https://github.com/graphql-go/graphql)
