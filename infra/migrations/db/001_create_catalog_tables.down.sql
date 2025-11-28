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