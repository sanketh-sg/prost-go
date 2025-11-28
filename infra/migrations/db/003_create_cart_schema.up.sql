-- Create carts table
CREATE TABLE IF NOT EXISTS cart.carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    total DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    abandoned_at TIMESTAMP NULL
);

-- Create cart items table
CREATE TABLE IF NOT EXISTS cart.cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES cart.carts(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create inventory locks table
CREATE TABLE IF NOT EXISTS cart.inventory_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES cart.carts(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    reservation_id UUID NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'locked',
    locked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    released_at TIMESTAMP NULL
);

-- Create saga state table
CREATE TABLE IF NOT EXISTS cart.saga_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id UUID NOT NULL UNIQUE,
    saga_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    cart_id UUID NOT NULL REFERENCES cart.carts(id),
    payload JSONB NOT NULL,
    compensation_log JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours'
);

-- Create idempotency records table
CREATE TABLE IF NOT EXISTS cart.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- Create indexes
CREATE INDEX idx_carts_user_id ON cart.carts(user_id);
CREATE INDEX idx_carts_status ON cart.carts(status);
CREATE INDEX idx_carts_created_at ON cart.carts(created_at);
CREATE INDEX idx_cart_items_cart_id ON cart.cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON cart.cart_items(product_id);
CREATE INDEX idx_inventory_locks_cart_id ON cart.inventory_locks(cart_id);
CREATE INDEX idx_inventory_locks_product_id ON cart.inventory_locks(product_id);
CREATE INDEX idx_inventory_locks_status ON cart.inventory_locks(status);
CREATE INDEX idx_inventory_locks_expires_at ON cart.inventory_locks(expires_at);
CREATE INDEX idx_saga_states_correlation_id ON cart.saga_states(correlation_id);
CREATE INDEX idx_saga_states_status ON cart.saga_states(status);
CREATE INDEX idx_saga_states_expires_at ON cart.saga_states(expires_at);
CREATE INDEX idx_idempotency_records_event_id ON cart.idempotency_records(event_id, service_name);