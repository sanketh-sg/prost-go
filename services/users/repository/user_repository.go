package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/sanketh-sg/prost/services/users/models"
    "github.com/sanketh-sg/prost/shared/db"
)

// UserRepository handles user database operations
type UserRepository struct {
	dbConn *db.Connection
}

// NewUserRepository creates a new user repository
func NewUserRepository(dbConn *db.Connection) *UserRepository {
	return &UserRepository{
		dbConn: dbConn,
	}
}

// CreateUser creates a new user in the database
func (userRepo *UserRepository) CreateUser(ctx context.Context, user *models.User) error{
	query := `
        INSERT INTO $schema.users (id, email, username, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, email, username, created_at, updated_at
    `
	query = replaceSchema(query, userRepo.dbConn.Schema)

	err := userRepo.dbConn.QueryRowContext(ctx, query, 
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID,&user.Email,&user.Username,&user.CreatedAt,&user.UpdatedAt) //copies the matched row to dest

    if err != nil {
        log.Printf("Error creating user: %v", err)
        return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (userRepo *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
	 	SELECT id, email, username, password_hash, created_at, updated_at, deleted_at
        FROM $schema.users
        WHERE email = $1 AND deleted_at IS NULL
	`

	query = replaceSchema(query, userRepo.dbConn.Schema)

	user := &models.User{}
	err := userRepo.dbConn.QueryRowContext(ctx, query, email).Scan(
        &user.ID,
        &user.Email,
        &user.Username,
        &user.PasswordHash,
        &user.CreatedAt,
        &user.UpdatedAt,
        &user.DeletedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }

    return user, nil

}

// GetUserByID retrieves a user by ID
func (userRepo *UserRepository) GetUserByID(ctx context.Context, userId string)(*models.User, error){
	query := ` 
		SELECT id, email, username, password_hash, created_at, updated_at, deleted_at
        FROM $schema.users
        WHERE id = $1 AND deleted_at IS NULL
	`
	query = replaceSchema(query,userRepo.dbConn.Schema)

	user := &models.User{}
	err := userRepo.dbConn.QueryRowContext(ctx,query,userId).Scan(
		&user.ID,
        &user.Email,
        &user.Username,
        &user.PasswordHash,
        &user.CreatedAt,
        &user.UpdatedAt,
        &user.DeletedAt,
	)
	if err != nil {
        return nil, fmt.Errorf("failed to get user by id: %w", err)
    }

    return user, nil
}
// UpdateUser updates user profile information
func (userRepo *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
    query := `
        UPDATE $schema.users
        SET email = $1, username = $2, updated_at = $3
        WHERE id = $4 AND deleted_at IS NULL
        RETURNING id, email, username, created_at, updated_at
    `

    query = replaceSchema(query, userRepo.dbConn.Schema)

    err := userRepo.dbConn.QueryRowContext(ctx, query,
        user.Email,
        user.Username,
        time.Now().UTC(),
        user.ID,
    ).Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }

    return nil
}
// DeleteUser soft deletes a user
func (userRepo *UserRepository) DeleteUser(ctx context.Context, id string) error {
    query := `
        UPDATE $schema.users
        SET deleted_at = $1, updated_at = $2
        WHERE id = $3
    `

    query = replaceSchema(query, userRepo.dbConn.Schema)

    result, err := userRepo.dbConn.ExecContext(ctx, query, time.Now().UTC(), time.Now().UTC(), id)
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }

    return nil
}
// EmailExists checks if email already exists
func (userRepo *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 FROM $schema.users 
            WHERE email = $1 AND deleted_at IS NULL
        )
    `

    query = replaceSchema(query, userRepo.dbConn.Schema)

    var exists bool
    err := userRepo.dbConn.QueryRowContext(ctx, query, email).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check email existence: %w", err)
    }

    return exists, nil
}
// UsernameExists checks if username already exists
func (userRepo *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
    query := `
        SELECT EXISTS(
            SELECT 1 FROM $schema.users 
            WHERE username = $1 AND deleted_at IS NULL
        )
    `

    query = replaceSchema(query, userRepo.dbConn.Schema)

    var exists bool
    err := userRepo.dbConn.QueryRowContext(ctx, query, username).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check username existence: %w", err)
    }

    return exists, nil
}
// Helper function to replace schema placeholder
func replaceSchema(query, schema string) string {
    for i := 0; i < len(query)-len("$schema"); i++ {
        if query[i:i+len("$schema")] == "$schema" {
            query = query[:i] + schema + query[i+len("$schema"):]
        }
    }
    return query
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string)(string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}
// VerifyPassword checks if the password matches the hash
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(password))

	return err == nil
}
