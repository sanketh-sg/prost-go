-- Drop indexes first
DROP INDEX IF EXISTS users.idx_oauth_providers_provider_sub;
DROP INDEX IF EXISTS users.idx_oauth_providers_user_id;
DROP INDEX IF EXISTS users.idx_idempotency_records_created_at;
DROP INDEX IF EXISTS users.idx_idempotency_records_event_id;
DROP INDEX IF EXISTS users.idx_users_created_at;
DROP INDEX IF EXISTS users.idx_users_username;
DROP INDEX IF EXISTS users.idx_users_email;

-- Drop tables 
DROP TABLE IF EXISTS users.oauth_providers CASCADE;
DROP TABLE IF EXISTS users.idempotency_records CASCADE;
DROP TABLE IF EXISTS users.users CASCADE;