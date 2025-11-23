-- +goose Up
-- +goose StatementBegin

-- Create catalog schema
CREATE SCHEMA IF NOT EXISTS catalog;

-- Create categories table
CREATE TABLE IF NOT EXISTS catalog.categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create products table
CREATE TABLE IF NOT EXISTS catalog.products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category_id BIGINT REFERENCES catalog.categories(id) ON DELETE SET NULL,
    sku VARCHAR(100) NOT NULL UNIQUE,
    stock_quantity INT NOT NULL DEFAULT 0,
    image_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create inventory reservations table (for saga state)
CREATE TABLE IF NOT EXISTS catalog.inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id BIGINT NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    order_id BIGINT NOT NULL,
    reservation_id UUID NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'reserved', -- reserved, released, expired
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '24 hours',
    released_at TIMESTAMP NULL
);

-- Create idempotency records table
CREATE TABLE IF NOT EXISTS catalog.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- Create indexes for better query performance
CREATE INDEX idx_products_category_id ON catalog.products(category_id);
CREATE INDEX idx_products_sku ON catalog.products(sku);
CREATE INDEX idx_products_created_at ON catalog.products(created_at);
CREATE INDEX idx_products_stock ON catalog.products(stock_quantity);
CREATE INDEX idx_categories_name ON catalog.categories(name);
CREATE INDEX idx_inventory_reservations_product_id ON catalog.inventory_reservations(product_id);
CREATE INDEX idx_inventory_reservations_order_id ON catalog.inventory_reservations(order_id);
CREATE INDEX idx_inventory_reservations_status ON catalog.inventory_reservations(status);
CREATE INDEX idx_inventory_reservations_expires_at ON catalog.inventory_reservations(expires_at);
CREATE INDEX idx_idempotency_records_event_id ON catalog.idempotency_records(event_id, service_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS catalog.idx_idempotency_records_event_id;
DROP INDEX IF EXISTS catalog.idx_inventory_reservations_expires_at;
DROP INDEX IF EXISTS catalog.idx_inventory_reservations_status;
DROP INDEX IF EXISTS catalog.idx_inventory_reservations_order_id;
DROP INDEX IF EXISTS catalog.idx_inventory_reservations_product_id;
DROP INDEX IF EXISTS catalog.idx_products_stock;
DROP INDEX IF EXISTS catalog.idx_products_created_at;
DROP INDEX IF EXISTS catalog.idx_products_sku;
DROP INDEX IF EXISTS catalog.idx_products_category_id;
DROP INDEX IF EXISTS catalog.idx_categories_name;

DROP TABLE IF EXISTS catalog.idempotency_records;
DROP TABLE IF EXISTS catalog.inventory_reservations;
DROP TABLE IF EXISTS catalog.products;
DROP TABLE IF EXISTS catalog.categories;

DROP SCHEMA IF EXISTS catalog;

-- +goose StatementEnd