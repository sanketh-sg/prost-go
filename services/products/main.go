package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
    "fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sanketh-sg/prost/services/products/handlers"
	"github.com/sanketh-sg/prost/services/products/middleware"
	"github.com/sanketh-sg/prost/services/products/repository"
	"github.com/sanketh-sg/prost/shared/db"
	"github.com/sanketh-sg/prost/shared/messaging"
)

func main()  {
	//Load env variables

    err := godotenv.Load(".env")

    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == ""{
        log.Println("Using default service name...")
		serviceName = "products"
	}

	port := os.Getenv("PORT_PRODUCT")
	if port == "" {
        log.Println("Using default port...")
		port = "8080"
	}

	dbSchema := os.Getenv("DB_SCHEMA")
	if dbSchema == "" {
        log.Println("Using default schema...")
		dbSchema = "catalog"
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
    if rabbitmqURL == "" {
        log.Panicln("Using default rabbitmqURL")
        rabbitmqURL = "amqp://guest:guest@localhost:5672/"
    }

	// Set Gin Mode
	// gin.SetMode(gin.ReleaseMode) // Disables debug logging, colorised output, better and faster

	log.Println("=== Products Service Starting ===")
    log.Printf("Service: %s", serviceName)
    log.Printf("Port: %s", port)
    log.Printf("Schema: %s", dbSchema)

	// DB Connection
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
	log.Println("Product-->Database connected")


	//RabbitMQ connection
	log.Println("\nConnecting to RabbitMQ...")
	rmqConn, err := messaging.NewRmqConnection(rabbitmqURL)
	if err != nil {
        log.Fatalf("RabbitMQ connection failed: %v", err)
    }
	defer rmqConn.Close()

	//Setup RabbitMQ Topology
	topology := messaging.GetProstTopology()
	if err := rmqConn.SetupRabbitMQ(topology); err != nil{
		log.Fatalf("RabbitMQ setup failed: %v", err)
	}
	log.Println("RabbitMQ connected and topology ready")

    // Initialize repositories
    productRepo := repository.NewProductRepository(dbConn)
    categoryRepo := repository.NewCategoryRepository(dbConn)
    inventoryRepo := repository.NewInventoryReservationRepository(dbConn)
    idempotencyStore := db.NewIdempotencyStore(dbConn)

	// Initialize event publisher
	publisher := messaging.NewPublisher(rmqConn, "products.events")

	// Initialize event subscriber
	subscriber := messaging.NewSubscriber(rmqConn, "products.events.queue")

	// Initialize handlers
    productHandler := handlers.NewProductHandler(
        productRepo,
        categoryRepo,
        inventoryRepo,
        idempotencyStore,
        publisher,
    )

	// Create Gin router
    router := gin.New()

	//Add Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Public routes
    router.GET("/health", productHandler.Health)
    router.GET("/categories", productHandler.GetCategories)
    router.GET("/categories/:id", productHandler.GetCategory)
    router.GET("/products", productHandler.GetProducts)
    router.GET("/products/:id", productHandler.GetProduct)

	// Admin routes (TODO: add authentication middleware in gateway)
    router.POST("/products", productHandler.CreateProduct)
    router.PATCH("/products/:id", productHandler.UpdateProduct)
    router.DELETE("/products/:id", productHandler.DeleteProduct)
    router.POST("/categories", productHandler.CreateCategory)

    // Inventory routes
    router.GET("/inventory/:product_id", productHandler.GetInventory)
    router.POST("/inventory/reserve", productHandler.ReserveInventory)
    router.POST("/inventory/release", productHandler.ReleaseInventory)


	// Server setup
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
	 // Start event subscriber in goroutine
    log.Println("\nStarting event subscriber...")
    
    go func() {
        log.Println("\nStarting event subscriber for inventory updates...")
        
        // Define the handler for inventory update events
        handler := func(message []byte) error {
            log.Printf("Processing inventory event: %s", string(message))
            
            // Parse the event
            event, err := subscriber.ParseEvent(message)
            if err != nil {
                return fmt.Errorf("failed to parse event: %w", err)
            }
            
            // Handle the event based on type
            // For now, just log it
            log.Printf("Event received: %v", event)
            return nil
        }
        
        // Subscribe with retry logic
        if err := subscriber.SubscribeWithRetry(handler, 3); err != nil {
            log.Fatalf("Subscriber error: %v", err)
        }
    }()

    // Start server in goroutine
    log.Printf("\n Products service listening on :%s", port)
    log.Println("\n=== Service Ready ===")

    _ = subscriber // Keep reference to prevent GC

    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }

    log.Println("✓ Service stopped")
}