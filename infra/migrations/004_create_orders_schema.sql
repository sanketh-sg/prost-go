-- +goose Up
-- +goose StatementBegin

-- Create orders schema
CREATE SCHEMA IF NOT EXISTS orders;

-- Create orders table
CREATE TABLE IF NOT EXISTS orders.orders (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    cart_id UUID NOT NULL,
    total DECIMAL(12, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, confirmed, shipped, delivered, cancelled
    saga_correlation_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    shipped_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL
);

-- Create order items table
CREATE TABLE IF NOT EXISTS orders.order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL, -- Price at time of purchase
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create saga state table for order processing saga
CREATE TABLE IF NOT EXISTS orders.saga_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id UUID NOT NULL UNIQUE,
    saga_type VARCHAR(100) NOT NULL, -- order_creation_saga, etc.
    status VARCHAR(50) NOT NULL DEFAULT 'pending', 
    -- States: pending -> cart_validated -> inventory_reserved -> confirmed (or failed -> compensated)
    order_id BIGINT REFERENCES orders.orders(id) ON DELETE SET NULL,
    payload JSONB NOT NULL, -- Store saga context
    compensation_log JSONB NOT NULL DEFAULT '[]', -- Track what needs to be rolled back
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours'
);

-- Create inventory reservations table (copy from cart for order fulfillment)
CREATE TABLE IF NOT EXISTS orders.inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id BIGINT NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    reservation_id UUID NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'reserved', -- reserved, released, expired, fulfilled
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours',
    released_at TIMESTAMP NULL,
    fulfilled_at TIMESTAMP NULL
);

-- Create compensation log table (tracks saga compensations)
CREATE TABLE IF NOT EXISTS orders.compensation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id BIGINT NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    saga_correlation_id UUID NOT NULL,
    compensation_event VARCHAR(100) NOT NULL, -- StockReleased, CartCleared, etc.
    compensation_payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, completed, failed
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL
);

-- Create idempotency records table for orders service
CREATE TABLE IF NOT EXISTS orders.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- Create indexes for performance
CREATE INDEX idx_orders_user_id ON orders.orders(user_id);
CREATE INDEX idx_orders_status ON orders.orders(status);
CREATE INDEX idx_orders_saga_correlation_id ON orders.orders(saga_correlation_id);
CREATE INDEX idx_orders_created_at ON orders.orders(created_at);
CREATE INDEX idx_order_items_order_id ON orders.order_items(order_id);
CREATE INDEX idx_order_items_product_id ON orders.order_items(product_id);
CREATE INDEX idx_saga_states_correlation_id ON orders.saga_states(correlation_id);
CREATE INDEX idx_saga_states_status ON orders.saga_states(status);
CREATE INDEX idx_saga_states_expires_at ON orders.saga_states(expires_at);
CREATE INDEX idx_inventory_reservations_order_id ON orders.inventory_reservations(order_id);
CREATE INDEX idx_inventory_reservations_status ON orders.inventory_reservations(status);
CREATE INDEX idx_compensation_log_order_id ON orders.compensation_log(order_id);
CREATE INDEX idx_compensation_log_correlation_id ON orders.compensation_log(saga_correlation_id);
CREATE INDEX idx_compensation_log_status ON orders.compensation_log(status);
CREATE INDEX idx_idempotency_records_event_id ON orders.idempotency_records(event_id, service_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS orders.idx_idempotency_records_event_id;
DROP INDEX IF EXISTS orders.idx_compensation_log_status;
DROP INDEX IF EXISTS orders.idx_compensation_log_correlation_id;
DROP INDEX IF EXISTS orders.idx_compensation_log_order_id;
DROP INDEX IF EXISTS orders.idx_inventory_reservations_status;
DROP INDEX IF EXISTS orders.idx_inventory_reservations_order_id;
DROP INDEX IF EXISTS orders.idx_saga_states_expires_at;
DROP INDEX IF EXISTS orders.idx_saga_states_status;
DROP INDEX IF EXISTS orders.idx_saga_states_correlation_id;
DROP INDEX IF EXISTS orders.idx_order_items_product_id;
DROP INDEX IF EXISTS orders.idx_order_items_order_id;
DROP INDEX IF EXISTS orders.idx_orders_created_at;
DROP INDEX IF EXISTS orders.idx_orders_saga_correlation_id;
DROP INDEX IF EXISTS orders.idx_orders_status;
DROP INDEX IF EXISTS orders.idx_orders_user_id;

DROP TABLE IF EXISTS orders.idempotency_records;
DROP TABLE IF EXISTS orders.compensation_log;
DROP TABLE IF EXISTS orders.inventory_reservations;
DROP TABLE IF EXISTS orders.saga_states;
DROP TABLE IF EXISTS orders.order_items;
DROP TABLE IF EXISTS orders.orders;

DROP SCHEMA IF EXISTS orders;

-- +goose StatementEnd