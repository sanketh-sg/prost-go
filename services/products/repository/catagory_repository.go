package repository

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/sanketh-sg/prost/services/products/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// CategoryRepository handles category database operations
type CategoryRepository struct {
    conn *db.Connection
}

// NewCategoryRepository creates new category repository
func NewCategoryRepository(conn *db.Connection) *CategoryRepository {
    return &CategoryRepository{conn: conn}
}

// CreateCategory creates a new category
func (cr *CategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
    query := `
        INSERT INTO $schema.categories (name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, description, created_at, updated_at
    `

    query = replaceSchema(query, cr.conn.Schema)

    err := cr.conn.QueryRowContext(ctx, query,
        category.Name,
        category.Description,
        category.CreatedAt,
        category.UpdatedAt,
    ).Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

    if err != nil {
        log.Printf("Error creating category: %v", err)
        return fmt.Errorf("failed to create category: %w", err)
    }

    return nil
}

// GetCategory retrieves a category by ID
func (cr *CategoryRepository) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
    query := `
        SELECT id, name, description, created_at, updated_at, deleted_at
        FROM $schema.categories
        WHERE id = $1 AND deleted_at IS NULL
    `

    query = replaceSchema(query, cr.conn.Schema)

    category := &models.Category{}
    err := cr.conn.QueryRowContext(ctx, query, id).Scan(
        &category.ID,
        &category.Name,
        &category.Description,
        &category.CreatedAt,
        &category.UpdatedAt,
        &category.DeletedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get category: %w", err)
    }

    return category, nil
}

// GetAllCategories retrieves all categories
func (cr *CategoryRepository) GetAllCategories(ctx context.Context) ([]*models.Category, error) {
    query := `
        SELECT id, name, description, created_at, updated_at, deleted_at
        FROM $schema.categories
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
    `

    query = replaceSchema(query, cr.conn.Schema)

    rows, err := cr.conn.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to get categories: %w", err)
    }
    defer rows.Close()

    var categories []*models.Category
    for rows.Next() {
        category := &models.Category{}
        err := rows.Scan(
            &category.ID,
            &category.Name,
            &category.Description,
            &category.CreatedAt,
            &category.UpdatedAt,
            &category.DeletedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan category: %w", err)
        }
        categories = append(categories, category)
    }

    return categories, nil
}

// UpdateCategory updates a category
func (cr *CategoryRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
    query := `
        UPDATE $schema.categories
        SET name = $1, description = $2, updated_at = $3
        WHERE id = $4 AND deleted_at IS NULL
        RETURNING id, name, description, created_at, updated_at
    `

    query = replaceSchema(query, cr.conn.Schema)

    err := cr.conn.QueryRowContext(ctx, query,
        category.Name,
        category.Description,
        time.Now().UTC(),
        category.ID,
    ).Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to update category: %w", err)
    }

    return nil
}

// DeleteCategory soft deletes a category
func (cr *CategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
    query := `
        UPDATE $schema.categories
        SET deleted_at = $1
        WHERE id = $2
    `

    query = replaceSchema(query, cr.conn.Schema)

    result, err := cr.conn.ExecContext(ctx, query, time.Now().UTC(), id)
    if err != nil {
        return fmt.Errorf("failed to delete category: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("category not found")
    }

    return nil
}