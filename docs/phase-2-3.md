## Phase 2 & 3: Key Learnings Summary

### Phase 2: Database Schema & Migrations

#### 1. **Schema Isolation Architecture**
```
Single PostgreSQL Database "prost"
├── users schema
│   ├── users table
│   └── idempotency_records
├── catalog schema
│   ├── products
│   ├── categories
│   ├── inventory_reservations
│   └── idempotency_records
├── cart schema
│   ├── carts
│   ├── cart_items
│   ├── inventory_locks
│   ├── saga_states
│   └── idempotency_records
└── orders schema
    ├── orders
    ├── order_items
    ├── saga_states
    ├── inventory_reservations
    ├── compensation_log
    └── idempotency_records
```

**Key Learning:** Each service owns its schema. No cross-schema foreign keys. Data isolation = microservices independence.

---

#### 2. **Why Idempotency Tables Per Service?**
- Each service tracks which events it has already processed (by `event_id`)
- **Without this:** Retried messages cause duplicate state changes
- **With this:** Same event processed 100 times = same result as once
- Example: `(event_id, service_name)` unique constraint = single processing guarantee

---

#### 3. **Saga State Tracking Tables**
```sql
saga_states table tracks distributed transaction progress:
- correlation_id: Links all events in one transaction
- status: pending → step1_complete → step2_complete → confirmed
- payload: JSONB context data
- compensation_log: What to rollback if failure
```

**Key Learning:** Saga state is the "source of truth" for multi-step operations. If a service crashes mid-saga, we know exactly where it failed and what to compensate.

---

#### 4. **Compensation Log for Rollbacks**
```
OrderPlaced → StockReserved → PaymentProcessed → CONFIRMED

If PaymentFailed:
- Compensation: Release reserved stock
- Compensation: Revert order status
- All tracked in compensation_log
```

**Key Learning:** Distributed transactions fail. Plan for it. Track compensations so you can undo partial changes.

---

#### 5. **Inventory Reservation Pattern**
Three tables track inventory across saga:
- `catalog.inventory_reservations` — What product service reserved
- `cart.inventory_locks` — What cart service locked (1 hour expiry)
- `orders.inventory_reservations` — Final reservation for order

**Key Learning:** Inventory flows through multiple services. Each needs its own view. TTL (time-to-live) prevents stale locks.

---

#### 6. **Index Strategy**
Every table has indexes on:
- Foreign keys (join performance)
- Frequently searched fields (email, username, SKU)
- Status fields (filter pending orders)
- Timestamps (date range queries)
- Unique constraints (prevent duplicates)

**Key Learning:** Migrations without indexes = slow queries in production. Plan indexes from day 1.

---

### Phase 3: Users Service - REST API & Authentication

#### 1. **Synchronous vs Asynchronous Services**
```
Users Service: SYNCHRONOUS (REST)
- Registration: Direct response
- Login: Returns JWT immediately
- Profile: CRUD operations synchronous
- NO events published (for now)

Why? User auth is blocking operation. Must succeed/fail immediately.
Can't say "Your registration will complete eventually"
```

**Key Learning:** Not all services are async. Authentication is synchronous by nature.

---

#### 2. **JWT Token Architecture**
```go
Claims = {
    UserID: "uuid-123",
    Email: "user@example.com", 
    Username: "john",
    ExpiresAt: time.Now() + 24h,
    IssuedAt: time.Now(),
    Issuer: "prost-users-service"
}

Token = HMAC_SHA256(Claims, secret_key)
```

**Key Learning:** JWT is stateless. No database lookup on every request. But must be valid JSON, signed, and not expired.

---

#### 3. **Password Security Best Practices**
```
❌ Never store: password in database
✅ Always store: bcrypt hash of password

Registration: Hash password → Store hash
Login: Hash provided password → Compare with stored hash
If match → Token generated
```

**Key Learning:** Bcrypt is slow on purpose (prevents brute force). Default cost=10 = ~100ms per hash. That's okay.

---

#### 4. **Middleware for Authentication**
```go
Public routes:
POST /register    — Anyone can register
POST /login       — Anyone can login
GET /health       — Anyone can check health

Protected routes:
GET /profile/:id       — Requires JWT in Authorization header
PATCH /profile/:id     — Requires JWT, only own profile

AuthMiddleware:
- Extracts token from "Bearer <token>" header
- Validates signature
- Checks expiration
- Stores user info in context for handlers
```

**Key Learning:** Middleware handles cross-cutting concerns (auth). Keeps handlers clean.

---

#### 5. **Repository Pattern for Data Access**
```
Handler → Repository → Database

Handler: HTTP logic
Repository: SQL queries, password hashing, validation
Database: Connection pooling, transactions

Separation of concerns: Easy to test, swap databases, change queries
```

**Key Learning:** Don't mix HTTP and database logic. Repository abstracts database.

---

#### 6. **Error Response Standardization**
```go
type ErrorResponse struct {
    Error   string `json:"error"`       // e.g., "email already exists"
    Message string `json:"message"`    // Details
    Code    int    `json:"code"`       // HTTP status
}

Consistency matters for clients (frontend knows format)
```

**Key Learning:** Standardized error responses prevent frontend from guessing error types.

---

#### 7. **Gin Framework Advantages**
```
Standard Library → Gin

Routing:
net/http: r.URL.Path parsing, manual checking
Gin: c.Param("id") — automatic

Error Handling:
net/http: Manual status codes, headers
Gin: c.JSON(status, data) — cleaner

Middleware:
net/http: Manual chain of handlers
Gin: router.Use(middleware) — declarative

Performance:
net/http: ~3k req/s
Gin: ~45k req/s (15x faster)
```

**Key Learning:** Gin reduces boilerplate. Faster. Better for microservices.

---

### Combined Phase 2 & 3: Data Flow

```
User Registration Flow:
1. POST /register {email, username, password}
   ↓
2. Handler validates request (Gin binding)
   ↓
3. Repository checks email/username uniqueness
   ↓
4. Repository hashes password with bcrypt
   ↓
5. Repository inserts user into users schema
   ↓
6. Return 201 Created with user ID
   ↓
7. (No event published yet - Phase 4)

User Login Flow:
1. POST /login {email, password}
   ↓
2. Repository finds user by email
   ↓
3. Repository verifies password hash
   ↓
4. JWT Manager generates signed token (24h expiry)
   ↓
5. Return 200 with token + user info

Protected Profile Update:
1. PATCH /profile/{id} with JWT header
   ↓
2. AuthMiddleware validates token signature
   ↓
3. AuthMiddleware checks user_id in token matches ID in URL
   ↓
4. Handler updates profile in database
   ↓
5. Return 200 with updated profile
```

---

### Critical Concepts for Phase 4+

| Concept | Why Important | Phase Needed |
|---------|--------------|-------------|
| **Idempotency** | Prevents duplicate processing of retried events | 4 (Products) |
| **Saga State** | Tracks distributed transaction progress | 6 (Orders) |
| **Event Routing** | Topic exchanges route events to correct queues | 4 (Products) |
| **Compensation** | Rollback partial changes on failure | 6 (Orders) |
| **Schema Isolation** | Services never interfere with each other's data | All phases |
| **JWT Middleware** | Protects endpoints, propagates user context | 7 (Gateway) |

---

### What You Now Understand

✅ **Database design** for microservices (separate schemas, idempotency, saga state)  
✅ **Authentication** (JWT tokens, password hashing, middleware)  
✅ **Distributed transactions** (saga pattern foundation)  
✅ **Event-driven communication** (prepared via messaging layer)  
✅ **REST API design** (Gin framework, standardized responses)  
✅ **Data isolation** (each service owns its schema)  

---

### Ready for Phase 4?

**Products Service will add:**
- Event publishing (ProductCreated, StockReserved, etc.)
- Inventory management with reservations
- Saga participation (reserve stock when orders arrive)
- CRUD endpoints for products
- Stock tracking across services

Same Gin architecture, but now with async event flow.

**Should I provide Phase 4: Products Service?** 🚀