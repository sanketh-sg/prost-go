package handlers

import (
    "context"
    "errors"

    "github.com/sanketh-sg/prost/services/users/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
    CreateUserFunc     func(ctx context.Context, user *models.User) error
    GetUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
    GetUserByIDFunc    func(ctx context.Context, userID string) (*models.User, error)
    UpdateUserFunc     func(ctx context.Context, user *models.User) error
    EmailExistsFunc    func(ctx context.Context, email string) (bool, error)
    UsernameExistsFunc func(ctx context.Context, username string) (bool, error)
	DeleteUserFunc     func(ctx context.Context, id string) error

}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    if m.GetUserByEmailFunc != nil {
        return m.GetUserByEmailFunc(ctx, email)
    }
    return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
    if m.GetUserByIDFunc != nil {
        return m.GetUserByIDFunc(ctx, userID)
    }
    return nil, errors.New("user not found")
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
    if m.UpdateUserFunc != nil {
        return m.UpdateUserFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
    if m.EmailExistsFunc != nil {
        return m.EmailExistsFunc(ctx, email)
    }
    return false, nil
}

func (m *MockUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
    if m.UsernameExistsFunc != nil {
        return m.UsernameExistsFunc(ctx, username)
    }
    return false, nil
}

func(m *MockUserRepository) DeleteUser(ctx context.Context, id string)(error){
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}