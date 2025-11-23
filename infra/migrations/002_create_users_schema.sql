-- +goose Up
-- +goose StatementBegin

-- Create users schema
CREATE SCHEMA IF NOT EXISTS users;

-- Create users table
CREATE TABLE IF NOT EXISTS users.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create idempotency records table for users service
CREATE TABLE IF NOT EXISTS users.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users.users(email);
CREATE INDEX idx_users_username ON users.users(username);
CREATE INDEX idx_users_created_at ON users.users(created_at);
CREATE INDEX idx_idempotency_records_event_id ON users.idempotency_records(event_id, service_name);
CREATE INDEX idx_idempotency_records_created_at ON users.idempotency_records(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS users.idx_idempotency_records_created_at;
DROP INDEX IF EXISTS users.idx_idempotency_records_event_id;
DROP INDEX IF EXISTS users.idx_users_created_at;
DROP INDEX IF EXISTS users.idx_users_username;
DROP INDEX IF EXISTS users.idx_users_email;

DROP TABLE IF EXISTS users.idempotency_records;
DROP TABLE IF EXISTS users.users;

DROP SCHEMA IF EXISTS users;

-- +goose StatementEnd