package handlers

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/users/auth"
    "github.com/sanketh-sg/prost/services/users/models"
    "github.com/sanketh-sg/prost/services/users/repository"

)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
    userRepo         repository.UserRepositoryInterface // Takes any implementation of UserRepositoryInterface
    jwtManager       *auth.JWTManager
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo repository.UserRepositoryInterface,jwtSecret string,) *UserHandler {
    return &UserHandler{
        userRepo:         userRepo,
        jwtManager:       auth.NewJWTManager(jwtSecret),
    }
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /register [post]
func (uh *UserHandler) Register(c *gin.Context) {
    // ctx := context.Background() // No timeout 
     ctx := c.Request.Context()  // Inherits HTTP server timeout

    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Validate request
    if valid, msg := req.Validate(); !valid {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation error",
            Message: msg,
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Check if email already exists
    exists, err := uh.userRepo.EmailExists(ctx, req.Email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "database error",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }
    if exists {
        c.JSON(http.StatusConflict, models.ErrorResponse{
            Error:   "email already exists",
            Message: "email exists",
            Code:    http.StatusConflict,
        })
        return
    }

    // Check if username already exists
    exists, err = uh.userRepo.UsernameExists(ctx, req.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "database error",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }
    if exists {
        c.JSON(http.StatusConflict, models.ErrorResponse{
            Error:   "username already exists",
            Message: "username exists",
            Code:    http.StatusConflict,
        })
        return
    }

    // Hash password
    passwordHash, err := repository.HashPassword(req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "password hashing failed",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    // Create user
    user := models.NewUser(req.Email, req.Username, passwordHash)
    if err := uh.userRepo.CreateUser(ctx, user); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to create user",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ User registered: %s (%s)", user.Email, user.ID)

    c.JSON(http.StatusCreated, gin.H{
        "message": "User registered successfully",
        "user": gin.H{
            "id":       user.ID,
            "email":    user.Email,
            "username": user.Username,
        },
    })
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /login [post]
func (uh *UserHandler) Login(c *gin.Context) {
    // ctx := context.Background()
     ctx := c.Request.Context()  // Inherits HTTP server timeout

    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Validate request
    if valid, msg := req.Validate(); !valid {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "validation error",
            Message: msg,
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get user by email
    user, err := uh.userRepo.GetUserByEmail(ctx, req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "invalid credentials",
            Message: "",
            Code:    http.StatusUnauthorized,
        })
        return
    }

    // Verify password
    if !repository.VerifyPassword(user.PasswordHash, req.Password) {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "invalid credentials",
            Message: "",
            Code:    http.StatusUnauthorized,
        })
        return
    }

    // Generate JWT token
    token, expiresAt, err := uh.jwtManager.GenerateToken(user.ID, user.Email, user.Username, 24*time.Hour)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "token generation failed",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ User logged in: %s", user.Email)

    c.JSON(http.StatusOK, models.LoginResponse{
        Token: token,
        User: models.User{
            ID:        user.ID,
            Email:     user.Email,
            Username:  user.Username,
            CreatedAt: user.CreatedAt,
            UpdatedAt: user.UpdatedAt,
        },
        ExpiresAt: expiresAt,
    })
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Retrieve user profile information (requires JWT)
// @Tags profile
// @Security Bearer
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /profile/{id} [get]
func (uh *UserHandler) GetProfile(c *gin.Context) {
    // ctx := context.Background()
     ctx := c.Request.Context()  // Inherits HTTP server timeout

    userID := c.Param("id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "user id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get user by ID
    user, err := uh.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "user not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":         user.ID,
        "email":      user.Email,
        "username":   user.Username,
        "created_at": user.CreatedAt,
        "updated_at": user.UpdatedAt,
    })
}

// UpdateProfile handles updating user profile
// @Summary Update user profile
// @Description Update user profile information (requires JWT)
// @Tags profile
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateProfileRequest true "Updated profile data"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Router /profile/{id} [patch]
func (uh *UserHandler) UpdateProfile(c *gin.Context) {
    // ctx := context.Background()
     ctx := c.Request.Context()  // Inherits HTTP server timeout

    userID := c.Param("id")
    if userID == "" {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "user id required",
            Message: "",
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get authenticated user ID from context
    authUserID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse{
            Error:   "user not authenticated",
            Message: "",
            Code:    http.StatusUnauthorized,
        })
        return
    }

    // Verify the token is for the same user
    if authUserID.(string) != userID {
        c.JSON(http.StatusForbidden, models.ErrorResponse{
            Error:   "cannot update other users",
            Message: "",
            Code:    http.StatusForbidden,
        })
        return
    }

    var req models.UpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error:   "invalid request body",
            Message: err.Error(),
            Code:    http.StatusBadRequest,
        })
        return
    }

    // Get current user
    user, err := uh.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse{
            Error:   "user not found",
            Message: err.Error(),
            Code:    http.StatusNotFound,
        })
        return
    }

    // Update fields if provided
    if req.Email != "" {
        user.Email = req.Email
    }
    if req.Username != "" {
        user.Username = req.Username
    }

    // Update user
    if err := uh.userRepo.UpdateUser(ctx, user); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error:   "failed to update user",
            Message: err.Error(),
            Code:    http.StatusInternalServerError,
        })
        return
    }

    log.Printf("✓ User profile updated: %s", userID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Profile updated successfully",
        "user": gin.H{
            "id":       user.ID,
            "email":    user.Email,
            "username": user.Username,
        },
    })
}

// Health handles health check
// @Summary Health check
// @Description Check service health
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (uh *UserHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "users",
        "time":    time.Now().UTC(),
    })
}