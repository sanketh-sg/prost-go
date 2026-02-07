-- Create users table
CREATE TABLE IF NOT EXISTS users.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create idempotency records table
CREATE TABLE IF NOT EXISTS users.idempotency_records (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    result VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, service_name)
);

-- OAuth providers linking table
CREATE TABLE IF NOT EXISTS users.oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_sub VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Composite unique constraint: one provider per user
    UNIQUE(user_id, provider)
);



-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users.users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users.users(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users.users(created_at);
CREATE INDEX IF NOT EXISTS idx_idempotency_records_event_id ON users.idempotency_records(event_id, service_name);
CREATE INDEX IF NOT EXISTS idx_idempotency_records_created_at ON users.idempotency_records(created_at);
CREATE INDEX IF NOT EXISTS idx_oauth_providers_user_id ON users.oauth_providers(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_providers_provider_sub ON users.oauth_providers(provider, provider_sub);