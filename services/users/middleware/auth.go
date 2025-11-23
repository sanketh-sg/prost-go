package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/users/auth"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    jwtManager := auth.NewJWTManager(jwtSecret)

    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "authorization header required",
            })
            c.Abort()
            return
        }

        // Extract token from "Bearer <token>"
        tokenString := authHeader
        if strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = authHeader[7:]
        }

        // Validate token
        claims, err := jwtManager.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "invalid token",
                "message": err.Error(),
            })
            c.Abort()
            return
        }

        // Store claims in context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("username", claims.Username)

        c.Next()
    }
}