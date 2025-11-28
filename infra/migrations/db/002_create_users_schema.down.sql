DROP INDEX IF EXISTS users.idx_idempotency_records_created_at;
DROP INDEX IF EXISTS users.idx_idempotency_records_event_id;
DROP INDEX IF EXISTS users.idx_users_created_at;
DROP INDEX IF EXISTS users.idx_users_username;
DROP INDEX IF EXISTS users.idx_users_email;

DROP TABLE IF EXISTS users.idempotency_records;
DROP TABLE IF EXISTS users.users;