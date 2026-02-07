package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/sanketh-sg/prost/services/users/handlers"
	"github.com/sanketh-sg/prost/services/users/middleware"
    "github.com/sanketh-sg/prost/services/users/auth"
	"github.com/sanketh-sg/prost/services/users/repository"
	"github.com/sanketh-sg/prost/shared/db"
)

func main() {
    
    err := godotenv.Load(".env")
	
    if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
    
	// Load environment variables
    serviceName := os.Getenv("SERVICE_NAME")
    if serviceName == "" {
        serviceName = "users"
    }

	port := os.Getenv("PORT_USER")
	if port == "" {
		port = "8083"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
        log.Fatal("DATABASE_URL not set")
    }

    dbSchema := os.Getenv("DB_SCHEMA")
    if dbSchema == "" {
        dbSchema = "users"
        log.Println("DATABASE_SCHEMA not set using default 'users'")
        
    }

	jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Println("JWT_SECRET not set, using default (INSECURE)")
        jwtSecret = "default-secret-change-in-production"
    }

    // Validate OAuth environment variables
    auth0Domain := os.Getenv("AUTH0_DOMAIN")
    auth0ClientID := os.Getenv("AUTH0_CLIENT_ID")
    auth0ClientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
    auth0RedirectURI := os.Getenv("AUTH0_REDIRECT_URI")

        if auth0Domain == "" || auth0ClientID == "" || auth0ClientSecret == "" || auth0RedirectURI == "" {
        log.Println("WARNING: OAuth environment variables not fully configured")
        log.Printf("   AUTH0_DOMAIN: %v", auth0Domain != "")
        log.Printf("   AUTH0_CLIENT_ID: %v", auth0ClientID != "")
        log.Printf("   AUTH0_CLIENT_SECRET: %v", auth0ClientSecret != "")
        log.Printf("   AUTH0_REDIRECT_URI: %v", auth0RedirectURI != "")
    }

	// Set Gin mode
    gin.SetMode(gin.ReleaseMode)  // Disables debug logging, colorised output, better and faster

	log.Println("=== Users Service Starting ===")
    log.Printf("Service: %s", serviceName)
    log.Printf("Port: %s", port)
    log.Printf("Schema: %s", dbSchema)
    log.Printf("Database URL: %s", dbURL)


	// Database connection
    log.Println("\nConnecting to PostgreSQL...")
    dbConn, err := db.NewDBConnection(db.Config{
        Host:     os.Getenv("HOST"),
        Port:     os.Getenv("PORT_DB"),
        User:     os.Getenv("USER"),
        Password: os.Getenv("PASSWORD"),
        DBName:   os.Getenv("DBNAME"),
        Schema:   dbSchema,
    })
    if err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer dbConn.DBConnClose()
    log.Println("✓ Database connected")


	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn)
    oauthProviderRepo := repository.NewOAuthProviderRepository(dbConn)

    // Initialize auth managers
    jwtManager := auth.NewJWTManager(jwtSecret)
    oauthManager := auth.NewOAuthManager()

    //Initialize Handlers
    userHandler := handlers.NewUserHandler(userRepo, jwtSecret)
    oauthHandler := handlers.NewOAuthHandler(oauthManager, jwtManager, oauthProviderRepo, userRepo)

	//Create Gin router
	router := gin.New()
	
	//Add Middleware
    router.Use(gin.Logger()) // Logs each request concurrently
    router.Use(gin.Recovery())  // Catches panics independently
    router.Use(middleware.CORSMiddleware()) // Takes care of CORS headers

	// Public routes
    router.POST("/register", userHandler.Register)
    router.POST("/login", userHandler.Login)
    router.GET("/health", userHandler.Health)

    // Public routes - OAuth (Auth0)
    router.GET("/oauth/login", oauthHandler.InitiateOAuth)
    router.GET("/oauth/login/gmail", oauthHandler.InitiateGmailOAuth)
    router.GET("/oauth/callback", oauthHandler.OAuthCallback)
    router.POST("/oauth/refresh", oauthHandler.RefreshToken)

	// Protected routes (require JWT)
    protected := router.Group("/")
    protected.Use(middleware.AuthMiddleware(jwtSecret))
    {
        protected.GET("profile/:id", userHandler.GetProfile)
        protected.PATCH("profile/:id", userHandler.UpdateProfile)
    }

	//Server Setup
	server := &http.Server{
		Addr:         ":" + port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
    log.Printf("\n Users service listening on :%s", port)
    log.Println("\n=== Service Ready ===")
	go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

	// Graceful shutdown
    sigChan := make(chan os.Signal, 1) // a channel to receive OS signals
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // listen for these signals like Ctrl+C or termination

    sig := <-sigChan // block until a signal is received
    log.Printf("\nReceived signal: %v", sig)
    log.Println("Shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // a timeout context for shutdown
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }

    log.Println("✓ Service stopped")

}

