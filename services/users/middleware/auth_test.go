package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/users/auth"
    "github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareSuccess(t *testing.T) {
    // Arrange
    jwtManager := auth.NewJWTManager("test-secret")
    token, _, _ := jwtManager.GenerateToken("user123", "test@example.com", "testuser", 1*time.Hour)

    // Create test router
    router := gin.New()
    router.Use(AuthMiddleware("test-secret"))
    router.GET("/test", func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"user_id": userID})
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "user123")
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
    // Arrange
    router := gin.New()
    router.Use(AuthMiddleware("test-secret"))
    router.GET("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "ok"})
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    // NO Authorization header
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "authorization header required")
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
    // Arrange
    router := gin.New()
    router.Use(AuthMiddleware("test-secret"))
    router.GET("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "ok"})
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token-xyz")
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
    // Arrange
    jwtManager := auth.NewJWTManager("test-secret")
    token, _, _ := jwtManager.GenerateToken("user123", "test@example.com", "testuser", -1*time.Hour) // Expired

    router := gin.New()
    router.Use(AuthMiddleware("test-secret"))
    router.GET("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "ok"})
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddlewareContextValues(t *testing.T) {
    // Arrange
    jwtManager := auth.NewJWTManager("test-secret")
    token, _, _ := jwtManager.GenerateToken("user123", "test@example.com", "testuser", 1*time.Hour)

    router := gin.New()
    router.Use(AuthMiddleware("test-secret"))
    router.GET("/test", func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        email, _ := c.Get("email")
        username, _ := c.Get("username")

        c.JSON(http.StatusOK, gin.H{
            "user_id":  userID,
            "email":    email,
            "username": username,
        })
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "user123")
    assert.Contains(t, w.Body.String(), "test@example.com")
    assert.Contains(t, w.Body.String(), "testuser")
}