-- Delete seed data (reverse order due to FK constraints)
-- DELETE FROM catalog.products WHERE sku IN ('LAPTOP-001', 'MOUSE-001', 'BOOK-GO-001', 'TSHIRT-001', 'JEANS-001');
-- DELETE FROM catalog.categories WHERE name IN ('Electronics', 'Books', 'Clothing');
DELETE FROM users.users WHERE email IN ('alice@example.com', 'bob@example.com', 'charlie@example.com');
