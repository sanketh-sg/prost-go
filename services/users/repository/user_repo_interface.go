package repository

import (
    "context"

    "github.com/sanketh-sg/prost/services/users/models"
)

// UserRepositoryInterface defines the contract for user repository operations
type UserRepositoryInterface interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    GetUserByID(ctx context.Context, userID string) (*models.User, error)
    UpdateUser(ctx context.Context, user *models.User) error
    DeleteUser(ctx context.Context, id string) error
    EmailExists(ctx context.Context, email string) (bool, error)
    UsernameExists(ctx context.Context, username string) (bool, error)
}
