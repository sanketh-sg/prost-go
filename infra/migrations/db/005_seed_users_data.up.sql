-- Seed sample users (password: password123)
INSERT INTO users.users (id, email, username, password_hash, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'alice@example.com', 'alice', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/1Cm', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'bob@example.com', 'bob', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/1Cm', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'charlie@example.com', 'charlie', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/1Cm', NOW(), NOW());

-- -- Seed sample products in catalog
-- INSERT INTO catalog.categories (name, description, created_at, updated_at) VALUES
-- ('Electronics', 'Electronic devices and gadgets', NOW(), NOW()),
-- ('Books', 'Physical and digital books', NOW(), NOW()),
-- ('Clothing', 'Apparel and accessories', NOW(), NOW());

-- INSERT INTO catalog.products (name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at) VALUES
-- ('Laptop Pro', 'High-performance laptop', 1299.99, 1, 'LAPTOP-001', 50, 'https://via.placeholder.com/300', NOW(), NOW()),
-- ('Wireless Mouse', 'Ergonomic wireless mouse', 29.99, 1, 'MOUSE-001', 200, 'https://via.placeholder.com/300', NOW(), NOW()),
-- ('The Go Programming Language', 'Learn Go from scratch', 49.99, 2, 'BOOK-GO-001', 100, 'https://via.placeholder.com/300', NOW(), NOW()),
-- ('T-Shirt', 'Cotton comfort t-shirt', 19.99, 3, 'TSHIRT-001', 300, 'https://via.placeholder.com/300', NOW(), NOW()),
-- ('Jeans', 'Classic blue jeans', 59.99, 3, 'JEANS-001', 150, 'https://via.placeholder.com/300', NOW(), NOW());