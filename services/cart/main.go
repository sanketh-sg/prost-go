package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sanketh-sg/prost/services/cart/handlers"
	"github.com/sanketh-sg/prost/services/cart/middleware"
	"github.com/sanketh-sg/prost/services/cart/repository"
	"github.com/sanketh-sg/prost/services/cart/subscribers"
	"github.com/sanketh-sg/prost/shared/db"
	"github.com/sanketh-sg/prost/shared/messaging"
)

func main() {
    // Load environment variables
    err := godotenv.Load(".env")

    if err != nil {
        log.Fatalln("Failed to load env file...")
    }

    serviceName := os.Getenv("SERVICE_NAME")
    if serviceName == "" {
        log.Println("Using default Service Name...")
        serviceName = "cart"
    }

    port := os.Getenv("PORT")
    if port == "" {
        log.Println("Using default port...")
        port = "8081"
    }

    dbSchema := os.Getenv("DB_SCHEMA")
    if dbSchema == "" {
        log.Println("Using default dbSchema...")
        dbSchema = "cart"
    }

    rabbitmqURL := os.Getenv("RABBITMQ_URL")
    if rabbitmqURL == "" {
        log.Panic("Using defalut RabbitMQ URL...")
        rabbitmqURL = "amqp://guest:guest@localhost:5672/"
    }

    // Set Gin mode
    // gin.SetMode(gin.ReleaseMode)

    log.Println("=== Cart Service Starting ===")
    log.Printf("Service: %s", serviceName)
    log.Printf("Port: %s", port)
    log.Printf("Schema: %s", dbSchema)

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

    // RabbitMQ connection
    log.Println("\nConnecting to RabbitMQ...")
    rmqConn, err := messaging.NewRmqConnection(rabbitmqURL)
    if err != nil {
        log.Fatalf("RabbitMQ connection failed: %v", err)
    }
    defer rmqConn.Close()

    // Setup RabbitMQ topology
    topology := messaging.GetProstTopology()
    if err := rmqConn.SetupRabbitMQ(topology); err != nil {
        log.Fatalf("RabbitMQ setup failed: %v", err)
    }
    log.Println("✓ RabbitMQ connected and topology ready")

    // Initialize repositories
    cartRepo := repository.NewCartRepository(dbConn)
    sagaRepo := repository.NewSagaStateRepository(dbConn)
    inventoryLockRepo := repository.NewInventoryLockRepository(dbConn)
    idempotencyStore := db.NewIdempotencyStore(dbConn)

    // Initialize event publisher (for cart.events exchange)
    publisher := messaging.NewPublisher(rmqConn, "cart.events")

    // Initialize event subscriber (listens to both cart.events and products.events)
    subscriber := messaging.NewSubscriber(rmqConn, "cart.events.queue")

    // Initialize handlers
    cartHandler := handlers.NewCartHandler(cartRepo, sagaRepo, inventoryLockRepo, idempotencyStore, publisher)

    // Create Gin router
    router := gin.New()

    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(middleware.CORSMiddleware())

    // Public routes
    router.GET("/health", cartHandler.Health)
    router.POST("/carts", cartHandler.CreateCart)
    router.GET("/carts", cartHandler.GetCart)
    router.POST("/carts/items", cartHandler.AddItem)
    router.DELETE("/carts/items/:product_id", cartHandler.RemoveItem)
    router.DELETE("/carts", cartHandler.DeleteCart)

    // Checkout endpoint (initiates saga)
    router.POST("/carts/checkout", cartHandler.CheckoutCart)

    // Server setup
    srv := &http.Server{
        Addr:         ":" + port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start event subscriber in background
    log.Println("\nStarting event subscriber...")
    go func() {
        eventHandler := subscribers.NewEventHandler(cartRepo, sagaRepo, inventoryLockRepo, idempotencyStore)
        if err := subscriber.Subscribe(func(message []byte) error {
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()
            return eventHandler.HandleEvent(ctx, message)
        }); err != nil {
            log.Printf("Subscriber error: %v", err)
        }
    }()

    // Start server in goroutine
    log.Printf("\n✓ Cart service listening on :%s", port)
    log.Println("\n=== Service Ready ===")

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    sig := <-sigChan
    log.Printf("\nReceived signal: %v", sig)
    log.Println("Shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }

    log.Println("✓ Service stopped")
}