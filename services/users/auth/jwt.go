package auth

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
    secret string
}

// Claims extends jwt.RegisteredClaims with custom claims
type Claims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    jwt.RegisteredClaims  // It includes standard claims like ExpiresAt, IssuedAt, etc.
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string) *JWTManager {
    return &JWTManager{secret: secret}
}

// GenerateToken generates a new JWT token with user claims and expiration
func (jm *JWTManager) GenerateToken(userID, email, username string, expiresIn time.Duration) (string, time.Time, error) {
    expiresAt := time.Now().UTC().Add(expiresIn)

    claims := Claims{
        UserID:   userID,
        Email:    email,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
            NotBefore: jwt.NewNumericDate(time.Now().UTC()),
            Issuer:    "prost-users-service",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(jm.secret))
    if err != nil {
        return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
    }

    return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jm.secret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return claims, nil
}