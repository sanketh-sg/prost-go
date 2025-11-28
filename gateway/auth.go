package main

import (
    "fmt"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

// UserClaims represents JWT claims
type UserClaims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// TokenValidator validates JWT tokens
type TokenValidator struct {
    secret string
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(secret string) *TokenValidator {
    return &TokenValidator{
        secret: secret,
    }
}

// ValidateToken validates and parses JWT token
func (tv *TokenValidator) ValidateToken(tokenString string) (*UserClaims, error) {
    // Remove "Bearer " prefix if present
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    claims := &UserClaims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(tv.secret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    if !token.Valid {
        return nil, fmt.Errorf("token is invalid")
    }

    return claims, nil
}