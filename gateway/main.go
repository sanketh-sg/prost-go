package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    // "github.com/sanketh-sg/prost/gateway"
)

// ContextKey is a custom type for context keys
type ContextKey string

const UserContextKey ContextKey = "user"

// Config holds gateway configuration
type Config struct {
    Port            string
    UsersServiceURL string
    ProductsServiceURL string
    CartServiceURL string
    OrdersServiceURL string
    JWTSecret string
}

// Gateway represents the API gateway
type Gateway struct {
    config *Config
    router *gin.Engine
    httpClient *HTTPClient
    tokenValidator *TokenValidator
}

// NewGateway creates a new gateway instance
func NewGateway(config *Config) *Gateway {
    return &Gateway{
        config: config,
        router: gin.Default(),
        httpClient: NewHTTPClient(),
        tokenValidator: NewTokenValidator(config.JWTSecret),
    }
}

// setupRoutes configures all gateway routes
func (g *Gateway) setupRoutes() {
    // CORS middleware
    g.router.Use(corsMiddleware())

    // Health check
    g.router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    // Build GraphQL schema
    schema := BuildSchema(g.httpClient, g.config)

    // Create service clients
    userService := NewUserService(g.config.UsersServiceURL, g.httpClient)
    productService := NewProductService(g.config.ProductsServiceURL, g.httpClient)
    cartService := NewCartService(g.config.CartServiceURL, g.httpClient)
    orderService := NewOrderService(g.config.OrdersServiceURL, g.httpClient)

    // Create resolver context
    resolverCtx := &ResolverContext{
        UserService:    userService,
        ProductService: productService,
        CartService:    cartService,
        OrderService:   orderService,
        TokenValidator: g.tokenValidator,
    }

    // Attach resolvers to schema
    AttachResolvers(schema, resolverCtx)

    // GraphQL endpoint
    g.router.POST("/graphql", authMiddleware(g.tokenValidator), func(c *gin.Context) {
        var query GraphQLQuery

        // Parse the JSON request body
        if err := c.BindJSON(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
        }
        
        // Create context with user claims
        ctx := c.Request.Context()
        if val, ok := c.Get("user"); ok {
            ctx = context.WithValue(ctx, UserContextKey, val)
        }

        // Create context with user claims
        // ctx := c.Request.Context()
        // if val, ok := c.Get("user"); ok {
        //     ctx = context.WithValue(ctx, "user", val)
        // }

        // Execute query
        result := ExecuteQuery(query.Query, query.Variables, schema, ctx)

        c.JSON(http.StatusOK, FormatResult(result))
    })

    // GraphQL introspection query
	g.router.GET("/graphql", func(c *gin.Context) {
		queryStr := c.Query("query")
		if queryStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
			return
		}

		result := ExecuteQuery(queryStr, nil, schema, c.Request.Context())
		c.JSON(http.StatusOK, FormatResult(result))
	})

    log.Println("✓ Routes configured")
}

// Run starts the gateway server
func (g *Gateway) Run() error {
    g.setupRoutes()

    // Create HTTP server with graceful shutdown
    server := &http.Server{
        Addr:    ":" + g.config.Port,
        Handler: g.router,
    }

    // Start server in background
    go func() {
        log.Printf("🚀 Gateway listening on port %s", g.config.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("❌ Server error: %v", err)
        }
    }()

    // Graceful shutdown on signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan

    log.Println("🛑 Shutting down gateway...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("⚠️  Shutdown error: %v", err)
        return err
    }

    log.Println("✓ Gateway stopped cleanly")
    return nil
}

// loadConfig loads configuration from environment
func loadConfig() *Config {
    // Load .env file if present
    err := godotenv.Load()

    if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

    port := os.Getenv("PORT")
    if port == "" {
        port = "80"
        log.Println("Using default port for gateway")
    }

    return &Config{
        Port: port,
        UsersServiceURL: os.Getenv("USERS_SERVICE_URL"),
        ProductsServiceURL: os.Getenv("PRODUCTS_SERVICE_URL"),
        OrdersServiceURL: os.Getenv("ORDERS_SERVICE_URL"),
        CartServiceURL: os.Getenv("CART_SERVICE_URL"),

        JWTSecret: os.Getenv("JWT_SECRET"),
    }
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}

// authMiddleware validates JWT token and extracts user claims
func authMiddleware(validator *TokenValidator) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            // Allow unauthenticated requests; resolvers will return error if auth required
            c.Next()
            return
        }

        claims, err := validator.ValidateToken(authHeader)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        c.Set("user", claims)
        c.Next()
    }
}

func main() {
    config := loadConfig()

    // Validate required config
    if config.UsersServiceURL == "" || config.ProductsServiceURL == "" ||
        config.CartServiceURL == "" || config.OrdersServiceURL == "" {
        log.Fatal("❌ Missing required service URLs in environment")
    }

    gateway := NewGateway(config)

    if err := gateway.Run(); err != nil {
        fmt.Printf("❌ Gateway error: %v\n", err)
        os.Exit(1)
    }
}