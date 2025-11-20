#!/bin/bash

# seed-data.sh - Seed the database with initial product data
# Usage: ./scripts/seed-data.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

DB_USER="prost_admin"
DB_PASSWORD="prost_password"
DB_NAME="prost"
DB_HOST="localhost"
DB_PORT="5432"

# Connection string
export PGPASSWORD="$DB_PASSWORD"

echo "🌱 Seeding Database with Sample Data..."
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "❌ Error: psql is not installed"
    echo "Please install PostgreSQL client tools"
    exit 1
fi

# Check database connection
echo "🔗 Checking database connection..."
if ! psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ Error: Cannot connect to database"
    echo "Make sure:"
    echo "  - PostgreSQL is running"
    echo "  - Docker containers are started (run: ./scripts/dev-start.sh)"
    exit 1
fi
echo "✅ Database connection successful"
echo ""

# Seed categories
echo "📂 Seeding categories..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" << EOF
INSERT INTO categories (name, description) VALUES
('Beer', 'Various types of beer beverages'),
('Wine', 'Red, white, and rosé wines'),
('Spirits', 'Liquor and distilled spirits'),
('Non-Alcoholic', 'Soft drinks and non-alcoholic beverages'),
('Energy Drinks', 'Energy and sports drinks')
ON CONFLICT (name) DO NOTHING;
EOF
echo "✅ Categories seeded"
echo ""

# Seed products
echo "🍺 Seeding products..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" << EOF
INSERT INTO products (name, description, price, category_id, sku, stock_quantity, image_url) 
SELECT 
    'Craft IPA',
    'Hoppy Indian Pale Ale with citrus notes',
    12.99,
    (SELECT id FROM categories WHERE name = 'Beer'),
    'BEER-IPA-001',
    150,
    'https://via.placeholder.com/200x300?text=Craft+IPA'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'BEER-IPA-001')
UNION ALL
SELECT 
    'Premium Lager',
    'Smooth and crisp pilsner-style lager',
    10.99,
    (SELECT id FROM categories WHERE name = 'Beer'),
    'BEER-LAGER-001',
    200,
    'https://via.placeholder.com/200x300?text=Lager'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'BEER-LAGER-001')
UNION ALL
SELECT 
    'Red Wine Blend',
    'Complex blend of Cabernet and Merlot',
    24.99,
    (SELECT id FROM categories WHERE name = 'Wine'),
    'WINE-RED-001',
    80,
    'https://via.placeholder.com/200x300?text=Red+Wine'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'WINE-RED-001')
UNION ALL
SELECT 
    'Chardonnay White',
    'Crisp and refreshing white wine',
    22.99,
    (SELECT id FROM categories WHERE name = 'Wine'),
    'WINE-WHITE-001',
    120,
    'https://via.placeholder.com/200x300?text=White+Wine'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'WINE-WHITE-001')
UNION ALL
SELECT 
    'Single Malt Whiskey',
    'Premium Scottish single malt whisky',
    49.99,
    (SELECT id FROM categories WHERE name = 'Spirits'),
    'SPIRIT-WHISKEY-001',
    50,
    'https://via.placeholder.com/200x300?text=Whiskey'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SPIRIT-WHISKEY-001')
UNION ALL
SELECT 
    'Vodka Premium',
    'Ultra-smooth premium vodka',
    39.99,
    (SELECT id FROM categories WHERE name = 'Spirits'),
    'SPIRIT-VODKA-001',
    75,
    'https://via.placeholder.com/200x300?text=Vodka'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SPIRIT-VODKA-001')
UNION ALL
SELECT 
    'Cola Classic',
    'Refreshing cola beverage',
    2.99,
    (SELECT id FROM categories WHERE name = 'Non-Alcoholic'),
    'SOFT-COLA-001',
    500,
    'https://via.placeholder.com/200x300?text=Cola'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SOFT-COLA-001')
UNION ALL
SELECT 
    'Orange Juice',
    'Fresh squeezed orange juice',
    3.99,
    (SELECT id FROM categories WHERE name = 'Non-Alcoholic'),
    'JUICE-OJ-001',
    300,
    'https://via.placeholder.com/200x300?text=Orange+Juice'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'JUICE-OJ-001')
UNION ALL
SELECT 
    'Energy Boost',
    'High-performance energy drink',
    4.99,
    (SELECT id FROM categories WHERE name = 'Energy Drinks'),
    'ENERGY-001',
    250,
    'https://via.placeholder.com/200x300?text=Energy+Drink'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'ENERGY-001')
UNION ALL
SELECT 
    'Sports Hydration',
    'Electrolyte-rich sports drink',
    3.49,
    (SELECT id FROM categories WHERE name = 'Energy Drinks'),
    'SPORTS-001',
    200,
    'https://via.placeholder.com/200x300?text=Sports+Drink'
WHERE NOT EXISTS (SELECT 1 FROM products WHERE sku = 'SPORTS-001');
EOF
echo "✅ Products seeded"
echo ""

# Display summary
echo "📊 Seeding Summary:"
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" << EOF
SELECT 
    'Categories' as "Entity",
    COUNT(*) as "Count"
FROM categories
UNION ALL
SELECT 
    'Products',
    COUNT(*)
FROM products;
EOF

echo ""
echo "✅ Database seeding completed!"
echo ""
echo "💡 You can now:"
echo "   - Browse products: GET http://localhost:8080/products"
echo "   - Get product details: GET http://localhost:8080/products/{id}"
