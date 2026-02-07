package models

import (
    "time"

    "github.com/google/uuid"
)

// User represents a user in the system
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    Username     string    `json:"username"`
    PasswordHash string    `json:"-"` // Never expose in JSON
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    *time.Time `json:"deleted_at,omitempty"`
    OAuthProviders []OAuthProvider `json:"oauth_providers,omitempty"`
}

// OAuthProvider represents an OAuth connection for a user
type OAuthProvider struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    Provider      string    `json:"provider"` // 'auth0', 'google', 'github'
    ProviderSub   string    `json:"-"`        // Never expose provider sub
    ProviderEmail string    `json:"provider_email,omitempty"`
    PictureURL    string    `json:"picture_url,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

// CreateUserRequest request body for user registration
type CreateUserRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// LoginRequest request body for user login
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// LoginResponse response containing JWT token
type LoginResponse struct {
    User         User   `json:"user"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
}
// UpdateProfileRequest request body for updating user profile
type UpdateProfileRequest struct {
    Email    string `json:"email,omitempty"`
    Username string `json:"username,omitempty"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}

// Validate validates CreateUserRequest
func (r CreateUserRequest) Validate() (bool, string) {
    if r.Email == "" {
        return false, "email is required"
    }
    if r.Username == "" {
        return false, "username is required"
    }
    if r.Password == "" {
        return false, "password is required"
    }
    if len(r.Password) < 6 {
        return false, "password must be at least 6 characters"
    }
    return true, ""
}

// Validate validates LoginRequest
func (r LoginRequest) Validate() (bool, string) {
    if r.Email == "" {
        return false, "email is required"
    }
    if r.Password == "" {
        return false, "password is required"
    }
    return true, ""
}

// NewUser creates a new user instance
func NewUser(email, username, passwordHash string) *User {
    now := time.Now().UTC()
    return &User{
        ID:           uuid.New().String(),
        Email:        email,
        Username:     username,
        PasswordHash: passwordHash,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
}