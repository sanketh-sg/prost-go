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


	// Set Gin mode
    gin.SetMode(gin.ReleaseMode)

	log.Println("=== Users Service Starting ===")
    log.Printf("Service: %s", serviceName)
    log.Printf("Port: %s", port)
    log.Printf("Schema: %s", dbSchema)
    log.Printf("Database URL: %s", dbURL)


	// Database connection
    log.Println("\nConnecting to PostgreSQL...")
    var host, envport, user, password, dbname string = os.Getenv("HOST"), os.Getenv("PORT_DB"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME")
    dbConn, err := db.NewDBConnection(db.Config{
        Host:     host,
        Port:     envport,
        User:     user,
        Password: password,
        DBName:   dbname,
        Schema:   dbSchema,
    })
    if err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer dbConn.DBConnClose()
    log.Println("✓ Database connected")


	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn)


	//Initialize Handlers
	userHandler := handlers.NewUserHandler(userRepo, jwtSecret)

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

