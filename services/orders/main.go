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
	"github.com/sanketh-sg/prost/services/orders/handlers"
	"github.com/sanketh-sg/prost/services/orders/middleware"
	"github.com/sanketh-sg/prost/services/orders/repository"
	"github.com/sanketh-sg/prost/services/orders/subscribers"
	"github.com/sanketh-sg/prost/shared/db"
	"github.com/sanketh-sg/prost/shared/messaging"
)

func main() {
    // Load environment variables

    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Failed to load env variables!")
    }

    serviceName := os.Getenv("SERVICE_NAME")
    if serviceName == "" {
        log.Println("Using Default service name...")
        serviceName = "orders"
    }

    port := os.Getenv("PORT")
    if port == "" {
        log.Println("Using Default port...")
        port = "8082"
    }

    dbSchema := os.Getenv("DB_SCHEMA")
    if dbSchema == "" {
        log.Println("Using Default schema name...")
        dbSchema = "orders"
    }

    rabbitmqURL := os.Getenv("RABBITMQ_URL")
    if rabbitmqURL == "" {
        log.Println("Using Default RabbitMQ URL...")
        rabbitmqURL = "amqp://guest:guest@localhost:5672/"
    }

    // Set Gin mode
    // gin.SetMode(gin.ReleaseMode)

    log.Println("=== Orders Service Starting ===")
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
        DBName:    os.Getenv("DBNAME"),
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
    orderRepo := repository.NewOrderRepository(dbConn)
    sagaRepo := repository.NewSagaStateRepository(dbConn)
    compensationRepo := repository.NewCompensationLogRepository(dbConn)
    inventoryResRepo := repository.NewInventoryReservationRepository(dbConn)
    idempotencyStore := db.NewIdempotencyStore(dbConn)

    // Initialize event publishers (for orders.events exchange)
    publisher := messaging.NewPublisher(rmqConn, "orders.events")

    // Initialize event subscriber (listens to cart.events and orders.events)
    subscriber := messaging.NewSubscriber(rmqConn, "orders.events.queue")

    // Initialize saga orchestrator
    sagaOrchestrator := subscribers.NewSagaOrchestrator(
        orderRepo,
        sagaRepo,
        compensationRepo,
        inventoryResRepo,
        idempotencyStore,
        publisher,
    )

    // Initialize handlers
    orderHandler := handlers.NewOrderHandler(
        orderRepo,
        sagaRepo,
        compensationRepo,
        inventoryResRepo,
        idempotencyStore,
        publisher,
        sagaOrchestrator,
    )

    // Create Gin router
    router := gin.New()

    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(middleware.CORSMiddleware())

    // Public routes
    router.GET("/health", orderHandler.Health)
    router.GET("/orders/:id", orderHandler.GetOrder)
    router.GET("/orders", orderHandler.GetOrders)

    // Saga routes
    router.GET("/sagas/:correlation_id", orderHandler.GetSagaState)
    router.POST("/orders/:id/cancel", orderHandler.CancelOrder)

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
        if err := subscriber.Subscribe(func(message []byte) error {
            return sagaOrchestrator.HandleEvent(context.Background(), message)
        }); err != nil {
            log.Printf("Subscriber error: %v", err)
        }
    }()

    // Start server in goroutine
    log.Printf("\n✓ Orders service listening on :%s", port)
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