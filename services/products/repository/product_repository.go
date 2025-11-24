package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/products/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// ProductRepository handles product database operations
type ProductRepository struct {
    conn *db.Connection
}

// NewProductRepository creates new product repository
func NewProductRepository(conn *db.Connection) *ProductRepository {
    return &ProductRepository{conn: conn}
}

// CreateProduct creates a new product
func (pr *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
    query := `
        INSERT INTO $schema.products 
        (name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at
    `

    query = replaceSchema(query, pr.conn.Schema)

    err := pr.conn.QueryRowContext(ctx, query,
        product.Name,
        product.Description,
        product.Price,
        product.CategoryID,
        product.SKU,
        product.StockQuantity,
        product.ImageURL,
        product.CreatedAt,
        product.UpdatedAt,
    ).Scan(
        &product.ID,
        &product.Name,
        &product.Description,
        &product.Price,
        &product.CategoryID,
        &product.SKU,
        &product.StockQuantity,
        &product.ImageURL,
        &product.CreatedAt,
        &product.UpdatedAt,
    )

    if err != nil {
        log.Printf("Error creating product: %v", err)
        return fmt.Errorf("failed to create product: %w", err)
    }

    return nil
}

// GetProduct retrieves a product by ID
func (pr *ProductRepository) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
    query := `
        SELECT id, name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at, deleted_at
        FROM $schema.products
        WHERE id = $1 AND deleted_at IS NULL
    `

    query = replaceSchema(query, pr.conn.Schema)

    product := &models.Product{}
    err := pr.conn.QueryRowContext(ctx, query, id).Scan(
        &product.ID,
        &product.Name,
        &product.Description,
        &product.Price,
        &product.CategoryID,
        &product.SKU,
        &product.StockQuantity,
        &product.ImageURL,
        &product.CreatedAt,
        &product.UpdatedAt,
        &product.DeletedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (pr *ProductRepository) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
    query := `
        SELECT id, name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at, deleted_at
        FROM $schema.products
        WHERE sku = $1 AND deleted_at IS NULL
    `

    query = replaceSchema(query, pr.conn.Schema)

    product := &models.Product{}
    err := pr.conn.QueryRowContext(ctx, query, sku).Scan(
        &product.ID,
        &product.Name,
        &product.Description,
        &product.Price,
        &product.CategoryID,
        &product.SKU,
        &product.StockQuantity,
        &product.ImageURL,
        &product.CreatedAt,
        &product.UpdatedAt,
        &product.DeletedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get product by sku: %w", err)
    }

    return product, nil
}

// GetAllProducts retrieves all products with optional category filter
func (pr *ProductRepository) GetAllProducts(ctx context.Context, categoryID *int64) ([]*models.Product, error) {
    query := `
        SELECT id, name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at, deleted_at
        FROM $schema.products
        WHERE deleted_at IS NULL
    `

    query = replaceSchema(query, pr.conn.Schema)

    var rows interface{}
    var err error

    if categoryID != nil {
        query += ` AND category_id = $1 ORDER BY created_at DESC`
        rows, err = pr.conn.QueryContext(ctx, query, *categoryID)
    } else {
        query += ` ORDER BY created_at DESC`
        rows, err = pr.conn.QueryContext(ctx, query)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to get products: %w", err)
    }

    return scanProducts(rows.(interface {
        Scan(...interface{}) error
        Next() bool
        Close() error
    }))
}

// UpdateProduct updates a product
func (pr *ProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
    query := `
        UPDATE $schema.products
        SET name = $1, description = $2, price = $3, stock_quantity = $4, image_url = $5, updated_at = $6
        WHERE id = $7 AND deleted_at IS NULL
        RETURNING id, name, description, price, category_id, sku, stock_quantity, image_url, created_at, updated_at
    `

    query = replaceSchema(query, pr.conn.Schema)

    err := pr.conn.QueryRowContext(ctx, query,
        product.Name,
        product.Description,
        product.Price,
        product.StockQuantity,
        product.ImageURL,
        time.Now().UTC(),
        product.ID,
    ).Scan(
        &product.ID,
        &product.Name,
        &product.Description,
        &product.Price,
        &product.CategoryID,
        &product.SKU,
        &product.StockQuantity,
        &product.ImageURL,
        &product.CreatedAt,
        &product.UpdatedAt,
    )

    if err != nil {
        return fmt.Errorf("failed to update product: %w", err)
    }

    return nil
}

// DeleteProduct soft deletes a product
func (pr *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
    query := `
        UPDATE $schema.products
        SET deleted_at = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, pr.conn.Schema)

    result, err := pr.conn.ExecContext(ctx, query, time.Now().UTC(), time.Now().UTC(), id)
    if err != nil {
        return fmt.Errorf("failed to delete product: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("product not found")
    }

    return nil
}

// DecrementStock decrements product stock
func (pr *ProductRepository) DecrementStock(ctx context.Context, productID int64, quantity int) error {
    query := `
        UPDATE $schema.products
        SET stock_quantity = stock_quantity - $1, updated_at = $2
        WHERE id = $3 AND stock_quantity >= $1 AND deleted_at IS NULL
    `

    query = replaceSchema(query, pr.conn.Schema)

    result, err := pr.conn.ExecContext(ctx, query, quantity, time.Now().UTC(), productID)
    if err != nil {
        return fmt.Errorf("failed to decrement stock: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("insufficient stock or product not found")
    }

    return nil
}

// IncrementStock increments product stock
func (pr *ProductRepository) IncrementStock(ctx context.Context, productID int64, quantity int) error {
    query := `
        UPDATE $schema.products
        SET stock_quantity = stock_quantity + $1, updated_at = $2
        WHERE id = $3 AND deleted_at IS NULL
    `

    query = replaceSchema(query, pr.conn.Schema)

    result, err := pr.conn.ExecContext(ctx, query, quantity, time.Now().UTC(), productID)
    if err != nil {
        return fmt.Errorf("failed to increment stock: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("product not found")
    }

    return nil
}

// Helper function
func replaceSchema(query, schema string) string {
    for i := 0; i < len(query)-len("$schema"); i++ {
        if query[i:i+len("$schema")] == "$schema" {
            query = query[:i] + schema + query[i+len("$schema"):]
        }
    }
    return query
}

func scanProducts(rows interface {
    Scan(...interface{}) error
    Next() bool
    Close() error
}) ([]*models.Product, error) {
    defer rows.Close()

    var products []*models.Product
    for rows.Next() {
        product := &models.Product{}
        err := rows.Scan(
            &product.ID,
            &product.Name,
            &product.Description,
            &product.Price,
            &product.CategoryID,
            &product.SKU,
            &product.StockQuantity,
            &product.ImageURL,
            &product.CreatedAt,
            &product.UpdatedAt,
            &product.DeletedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan product: %w", err)
        }
        products = append(products, product)
    }

    return products, nil
}