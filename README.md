# Prost! Beverage Store

A modern microservices-based e-commerce platform built with Vue 3 and Go, demonstrating enterprise-grade architecture with proper service boundaries, API gateway patterns, and cloud-native deployment strategies.

## 🏗️ Architecture Overview

This project implements a distributed microservices system for an online beverage store with the following service boundaries:

- **API Gateway** — Central ingress and request router
- **Auth Service** — User authentication and JWT token management
- **Catalog Service** — Product catalog and categories
- **Cart Service** — User shopping carts
- **Orders Service** — Order placement and lifecycle
- **Inventory Service** — Stock management and reservations
- **Payments Service** — Payment processing integration
- **Shipping Service** — Shipment tracking
- **Notifications Service** — Email and SMS notifications
- **Admin/Analytics Service** — Internal reporting

## 🛠️ Tech Stack

| Component | Technology | Notes |
|-----------|-----------|-------|
| **Backend** | Go | gRPC/HTTP/REST microservices |
| **Frontend** | Vue 3 + Vite | Modern reactive UI framework |
| **State Management** | Pinia | Vue store management |
| **Routing** | Vue Router | Frontend navigation |
| **Databases** | PostgreSQL (per service) | Database per service pattern |
| **Caching** | Redis | Sessions and cache layer |
| **File Storage** | MinIO | Object storage for images |
| **Message Queue** | Kafka / RabbitMQ | Async event processing |
| **Authentication** | OpenID Connect + JWT | Secure token-based auth |
| **Containerization** | Docker | Service packaging |
| **Orchestration** | Kubernetes | Production deployment (kind/k3s/EKS/GKE/AKS) |
| **Infrastructure as Code** | Terraform | Cloud infrastructure provisioning |
| **CI/CD** | GitHub Actions | Automated pipelines |
| **Docker Registry** | GHCR | Container image hosting |

## 📋 Service-by-Service Implementation Guide

### 1. API Gateway (Reverse Proxy)

**Purpose:** Central ingress point that routes all requests to appropriate microservices

**Responsibilities:**
- Route requests to downstream services
- Handle CORS and request compression (gzip)
- Request logging and tracing
- Rate limiting
- JWT middleware for authentication

**Implementation Tasks:**
- [ ] Configure gateway routes
  - `/api/catalog/*` → catalog-service
  - `/api/cart/*` → cart-service
  - `/api/orders/*` → orders-service
  - `/api/auth/*` → auth-service
- [ ] Implement CORS and compression middleware
- [ ] Add comprehensive request logging
- [ ] Implement rate limiting (Phase 2)
- [ ] Add JWT validation middleware (Phase 5)

---

### 2. Catalog Service

**Purpose:** Manage product catalog, categories, and image metadata

**Responsibilities:**
- Product and category management
- Product details and search
- Image metadata storage

**API Endpoints:**
- `GET /health` — Health check
- `GET /products` — List all products
- `GET /products/:id` — Get product details

**Implementation Tasks:**
- [ ] Scaffold Go service with router and health check
- [ ] Setup PostgreSQL connection and migrations
- [ ] Implement domain models and repository layer
- [ ] Develop product handlers (list, get by ID)
- [ ] Seed database with sample products
- [ ] Add MinIO integration for image storage (Phase 2)
- [ ] Add Redis caching layer (Phase 2)

## C. Cart Service

Purpose: Manage user shopping carts

**API Endpoints:**
- `GET /cart` — Retrieve current cart
- `POST /cart/add` — Add item to cart
- `POST /cart/remove` — Remove item from cart
- `DELETE /cart/clear` — Clear entire cart

**Implementation Tasks:**
- [ ] Scaffold Go service
- [ ] Implement in-memory cart store (Phase 1)
- [ ] Develop cart endpoints
- [ ] Add catalog service validation for product IDs
- [ ] Migrate to Redis storage (Phase 2)
- [ ] Add cart expiration with TTL (Phase 3)

## D. Orders Service

Purpose: Order placement and order history management

**API Endpoints:**
- `POST /orders` — Create new order
- `GET /orders/:id` — Get order details
- `GET /orders/user/:id` — Get user's order history

**Implementation Tasks:**
- [ ] Scaffold Go service
- [ ] Setup PostgreSQL with order schema migrations
- [ ] Implement order creation handler
- [ ] Integrate cart service for order items
- [ ] Develop order retrieval endpoints
- [ ] Call inventory service for stock checks (Phase 3)
- [ ] Emit order placement events to message queue (Phase 4)

## E. Auth Service

Purpose: User authentication and JWT token management

**API Endpoints:**
- `POST /auth/signup` — Register new user
- `POST /auth/login` — Authenticate and issue tokens
- `GET /auth/me` — Get current user profile

**Implementation Tasks:**
- [ ] Scaffold Go service
- [ ] Create user table migrations
- [ ] Implement password hashing (bcrypt/argon2)
- [ ] Develop signup handler with validation
- [ ] Develop login handler with JWT generation
- [ ] Implement refresh token mechanism
- [ ] Add JWT validation middleware to gateway

## F. Inventory Service

Purpose: Stock management and reservations

**API Endpoints:**
- `POST /inventory/reserve` — Reserve stock for order
- `POST /inventory/release` — Release reserved stock
- `POST /inventory/deduct` — Deduct stock on fulfillment

**Implementation Tasks:**
- [ ] Create SKU and stock tables
- [ ] Implement database transactions for race condition handling
- [ ] Develop reservation endpoints
- [ ] Consume order placement events from message queue
- [ ] Emit inventory reserved/deducted events

## G. Payments Service

Purpose: Payment processing and PSP integration

**API Endpoints:**
- `POST /payments/charge` — Process payment

**Implementation Tasks:**
- [ ] Create fake payment provider for testing
- [ ] Implement charge endpoint with idempotency keys
- [ ] Add transaction logging
- [ ] Integrate real payment provider (Stripe/PayPal) (Phase 3)

## H. Notifications Service

Purpose: Send transactional notifications

**Implementation Tasks:**
- [ ] Consume order confirmation events
- [ ] Setup SMTP provider integration
- [ ] Create notification templates
- [ ] Implement retry logic for failed deliveries
- [ ] Add SMS provider integration (Phase 3)

# 🔄 Infrastructure & DevOps

## Infrastructure Requirements

- [ ] Kubernetes manifests or Helm charts
- [ ] Dockerfiles for each service
- [ ] Local development environment (docker-compose or kind)
- [ ] Observability stack:
  - Prometheus (metrics collection)
  - Jaeger (distributed tracing)
  - Loki (log aggregation)

## CI/CD Pipeline

- [ ] GitHub Actions workflow setup
- [ ] Build automation for services
- [ ] Automated testing on PR
- [ ] Image scanning and security checks
- [ ] Automatic deployment on main branch

# 🎨 Frontend Implementation

## Vue 3 Application Structure

### Core Setup
- [ ] Initialize Vue 3 + Vite + Pinia + Vue Router
- [ ] Configure ESLint and code formatting
- [ ] Setup auto-import for components and composables
- [ ] Create global error handling and logging

### UI Layout
- [ ] Design responsive header component
- [ ] Create footer with links and information
- [ ] Implement consistent navigation menu
- [ ] Add responsive mobile menu

### Pages

#### Core Pages
- [ ] **Home** — Landing page with featured products
- [ ] **Product List** — Browse and filter products with pagination
- [ ] **Product Details** — Single product view with reviews/ratings
- [ ] **Shopping Cart** — Manage items, adjust quantities
- [ ] **Checkout** — Order review and confirmation
- [ ] **Order Confirmation** — Order summary and tracking info
- [ ] **Order History** — View past orders and details

#### Authentication Pages
- [ ] **Login** — User authentication form
- [ ] **Signup** — User registration form
- [ ] **User Profile** — Account settings and preferences

### State Management (Pinia Stores)

| Store | Responsibilities |
|-------|-----------------|
| `authStore` | User authentication state, JWT tokens, current user |
| `productStore` | Product catalog caching, search/filter state |
| `cartStore` | Shopping cart items, totals, quantity management |
| `orderStore` | Order history, order details, checkout state |

### Services & Integration

- [ ] **API Client** — Axios instance with base URL pointing to API gateway
  - Automatic JWT token injection in request headers
  - Request/response interceptors for error handling
  - Base URL configuration for different environments
- [ ] **Error Handling** — Global error handler with toast notifications
- [ ] **Route Guards** — Protect routes requiring authentication
- [ ] **Loading States** — Global loading indicator management

### UI Components

- [ ] Product cards with images and pricing
- [ ] Category filter sidebar
- [ ] Shopping cart item list with edit/remove actions
- [ ] Toast notification system
- [ ] Modal dialogs for confirmations
- [ ] Loading spinners and skeletons
- [ ] Form validation and error display

---

## 📊 Implementation Roadmap

### Phase 1: Foundation & Setup

**Goal:** Establish project structure and infrastructure

1. Setup monorepo structure (frontend, services, infra)
2. Configure Docker and local development environment (docker-compose)
3. Setup PostgreSQL, Redis, and MinIO instances
4. Create API gateway with basic routing
5. Initialize Vue application scaffold

**Deliverables:** All services run locally, basic gateway routing works

---

### Phase 2: First Vertical Slice — Product Catalog

**Goal:** End-to-end flow: Frontend → Gateway → Catalog Service → Database

1. Create Catalog service skeleton with health check
2. Setup PostgreSQL and create migrations for products/categories
3. Seed database with sample beverage products
4. Implement product endpoints (GET /products, GET /products/:id)
5. Expose catalog through API gateway routing
6. Build Vue product list page
7. Implement product detail view

**Deliverables:** Users can browse and view product details 🎉

---

### Phase 3: Second Vertical Slice — Shopping Cart

**Goal:** Add shopping cart functionality end-to-end

1. Create Cart service with in-memory store (Phase 1) or Redis (Phase 2)
2. Implement cart endpoints (GET, POST /cart/add, POST /cart/remove)
3. Add product validation by calling Catalog service
4. Build cart UI component and page
5. Implement add-to-cart and remove-from-cart actions
6. Display cart total and item count

**Deliverables:** Users can add/remove items and manage cart 🎉

---

### Phase 4: Third Vertical Slice — Orders

**Goal:** Complete shopping flow with order placement

1. Create Orders service with PostgreSQL
2. Implement POST /orders endpoint (fetch cart items, calculate totals, store order)
3. Integrate with Cart and Catalog services
4. Build checkout page in Vue
5. Implement order confirmation page with order summary
6. Add order tracking information

**Deliverables:** Users can complete purchases and view orders 🎉

---

### Phase 5: Authentication

**Goal:** Add user authentication and authorization

1. Create Auth service (signup, login, JWT generation)
2. Implement user table with secure password storage (bcrypt/argon2)
3. Add JWT validation middleware to API gateway
4. Build login and signup pages in Vue
5. Create user profile page
6. Implement route guards for protected pages
7. Add session management and token refresh

**Deliverables:** Multi-user system with secure authentication 🎉

---

### Phase 6: Advanced Features (Inventory & Payments)

**Goal:** Add real-world commerce features

1. Create Inventory service with stock tracking
2. Implement stock reservation on order placement
3. Add payment service with fake payment provider (testing)
4. Integrate inventory checks into checkout
5. Emit order events to message queue (Kafka/RabbitMQ)
6. Implement event consumers for inventory updates
7. Integrate real payment provider (Stripe/PayPal)

**Deliverables:** Stock-aware e-commerce with payment processing 🎉

---

### Phase 7: Observability & Production Deployment

**Goal:** Add monitoring, logging, and deploy to production

1. Implement structured logging across all services
2. Setup Prometheus for metrics collection
3. Configure Jaeger for distributed tracing
4. Setup Loki for log aggregation
5. Create Kubernetes manifests/Helm charts
6. Build GitHub Actions CI/CD pipeline
7. Implement automated testing and code scanning
8. Deploy to cloud (EKS/GKE/AKS) or on-prem Kubernetes

**Deliverables:** Production-ready, observable microservices architecture 🎉

---

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Compose
- Node.js 18+ (for frontend and tooling)
- Go 1.21+ (for backend services)
- kubectl (for Kubernetes deployment)

### Local Development Setup

```bash
# Clone repository
git clone <repo-url>
cd prost

# Start infrastructure (Docker containers)
docker-compose up -d

# Frontend setup
cd frontend
pnpm install
pnpm dev

# Backend services are auto-started in docker-compose
```

### Project Structure

```
prost/
├── frontend/           # Vue 3 application
│   ├── src/
│   ├── public/
│   └── package.json
├── services/           # Go microservices
│   ├── catalog/
│   ├── cart/
│   ├── orders/
│   ├── auth/
│   ├── inventory/
│   ├── payments/
│   ├── shipping/
│   └── notifications/
├── infra/             # Kubernetes & Docker configs
│   ├── k8s/
│   ├── docker-compose.yml
│   └── terraform/
└── README.md
```

---

## 📚 Documentation

- **[Frontend Docs](./frontend/README.md)** — Vue setup, components, stores
- **[Backend Docs](./services/README.md)** — Service APIs, architecture decisions
- **[Infrastructure Docs](./infra/README.md)** — Kubernetes, Docker, deployment guides

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Create a feature branch (`git checkout -b feature/amazing-feature`)
2. Commit changes (`git commit -m 'Add amazing feature'`)
3. Push to branch (`git push origin feature/amazing-feature`)
4. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

---

## 💡 Key Architectural Patterns

- **Microservices:** Each service has its own database (database per service pattern)
- **API Gateway:** Single entry point for all client requests
- **Event-Driven:** Async communication via message queues
- **Service-to-Service Communication:** HTTP/REST with circuit breakers
- **Security:** JWT tokens, RBAC, input validation
- **Observability:** Structured logging, distributed tracing, metrics

---

## 🎯 Next Steps

Start with **Phase 1** to setup your development environment, then proceed through phases sequentially to build a complete e-commerce platform. Each phase builds upon the previous one with a working vertical slice of functionality.