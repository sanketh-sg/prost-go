package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
	"errors"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/users/models"
    "github.com/sanketh-sg/prost/services/users/repository"
    "github.com/stretchr/testify/assert"
)

// ===== REGISTER TESTS =====

func TestRegisterSuccess(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
            return false, nil
        },
        UsernameExistsFunc: func(ctx context.Context, username string) (bool, error) {
            return false, nil
        },
        CreateUserFunc: func(ctx context.Context, user *models.User) error {
            return nil
        },
    }

    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder() // This is required to record HTTP responses
    c, _ := gin.CreateTestContext(w) // Create a Gin context for testing with the recorder

    payload := models.CreateUserRequest{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "User registered successfully", response["message"])
    assert.NotNil(t, response["user"])
}

func TestRegisterInvalidJSON(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "invalid request body", response.Error)
}

func TestRegisterMissingEmail(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.CreateUserRequest{
        Email:    "",
        Username: "testuser",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "validation error", response.Error)
    assert.Equal(t, "email is required", response.Message)
}

func TestRegisterPasswordTooShort(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.CreateUserRequest{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "12345", // Too short
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "password must be at least 6 characters", response.Message)
}

func TestRegisterDuplicateEmail(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
            return true, nil // Email already exists
        },
    }
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.CreateUserRequest{
        Email:    "existing@example.com",
        Username: "testuser",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusConflict, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "email already exists", response.Error)
}

func TestRegisterDuplicateUsername(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        EmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
            return false, nil
        },
        UsernameExistsFunc: func(ctx context.Context, username string) (bool, error) {
            return true, nil // Username already exists
        },
    }
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.CreateUserRequest{
        Email:    "test@example.com",
        Username: "existing",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Register(c)

    // Assert
    assert.Equal(t, http.StatusConflict, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "username already exists", response.Error)
}

// ===== LOGIN TESTS =====

func TestLoginSuccess(t *testing.T) {
    // Arrange
    hashedPassword, _ := repository.HashPassword("password123")
    mockUser := &models.User{
        ID:           "user123",
        Email:        "test@example.com",
        Username:     "testuser",
        PasswordHash: hashedPassword,
        CreatedAt:    time.Now().UTC(),
        UpdatedAt:    time.Now().UTC(),
    }

    mockRepo := &MockUserRepository{
        GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
            return mockUser, nil
        },
    }

    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    var response models.LoginResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotEmpty(t, response.AccessToken)
    assert.Equal(t, mockUser.Email, response.User.Email)
    assert.Equal(t, mockUser.ID, response.User.ID)
}

func TestLoginInvalidJSON(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "invalid request body", response.Error)
}

func TestLoginMissingEmail(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.LoginRequest{
        Email:    "",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "email is required", response.Message)
}

func TestLoginUserNotFound(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
            return nil, errors.New("user not found")
        },
    }
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.LoginRequest{
        Email:    "nonexistent@example.com",
        Password: "password123",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "invalid credentials", response.Error)
}

func TestLoginWrongPassword(t *testing.T) {
    // Arrange
    hashedPassword, _ := repository.HashPassword("correctpassword")
    mockUser := &models.User{
        ID:           "user123",
        Email:        "test@example.com",
        Username:     "testuser",
        PasswordHash: hashedPassword,
        CreatedAt:    time.Now().UTC(),
        UpdatedAt:    time.Now().UTC(),
    }

    mockRepo := &MockUserRepository{
        GetUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
            return mockUser, nil
        },
    }

    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    payload := models.LoginRequest{
        Email:    "test@example.com",
        Password: "wrongpassword",
    }
    body, _ := json.Marshal(payload)
    c.Request = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
    c.Request.Header.Set("Content-Type", "application/json")

    // Act
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "invalid credentials", response.Error)
}

// ===== GET PROFILE TESTS =====

func TestGetProfileSuccess(t *testing.T) {
    // Arrange
    mockUser := &models.User{
        ID:        "user123",
        Email:     "test@example.com",
        Username:  "testuser",
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }

    mockRepo := &MockUserRepository{
        GetUserByIDFunc: func(ctx context.Context, userID string) (*models.User, error) {
            return mockUser, nil
        },
    }

    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{gin.Param{Key: "id", Value: "user123"}}
    c.Request = httptest.NewRequest(http.MethodGet, "/profile/user123", nil)

    // Act
    handler.GetProfile(c)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "user123", response["id"])
    assert.Equal(t, "test@example.com", response["email"])
    assert.Equal(t, "testuser", response["username"])
}

func TestGetProfileMissingID(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest(http.MethodGet, "/profile/", nil)

    // Act
    handler.GetProfile(c)

    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "user id required", response.Error)
}

func TestGetProfileNotFound(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        GetUserByIDFunc: func(ctx context.Context, userID string) (*models.User, error) {
            return nil, errors.New("user not found")
        },
    }

    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{gin.Param{Key: "id", Value: "nonexistent"}}
    c.Request = httptest.NewRequest(http.MethodGet, "/profile/nonexistent", nil)

    // Act
    handler.GetProfile(c)

    // Assert
    assert.Equal(t, http.StatusNotFound, w.Code)
    var response models.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "user not found", response.Error)
}

// ===== HEALTH CHECK TEST =====

func TestHealth(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    handler := NewUserHandler(mockRepo, "test-secret")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

    // Act
    handler.Health(c)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "healthy", response["status"])
    assert.Equal(t, "users", response["service"])
}