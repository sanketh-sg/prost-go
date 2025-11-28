-- Create orders table
CREATE TABLE IF NOT EXISTS orders.orders (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    cart_id UUID NOT NULL,
    total DECIMAL(12, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
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
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create saga state table
CREATE TABLE IF NOT EXISTS orders.saga_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id UUID NOT NULL UNIQUE,
    saga_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    order_id BIGINT REFERENCES orders.orders(id) ON DELETE SET NULL,
    payload JSONB NOT NULL,
    compensation_log JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours'
);

-- Create inventory reservations table
CREATE TABLE IF NOT EXISTS orders.inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id BIGINT NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    reservation_id UUID NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'reserved',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours',
    released_at TIMESTAMP NULL,
    fulfilled_at TIMESTAMP NULL
);

-- Create compensation log table
CREATE TABLE IF NOT EXISTS orders.compensation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id BIGINT NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    saga_correlation_id UUID NOT NULL,
    compensation_event VARCHAR(100) NOT NULL,
    compensation_payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL
);

-- Create idempotency records table
CREATE TABLE IF NOT EXISTS orders.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- Create indexes
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