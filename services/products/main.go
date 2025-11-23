package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sanketh-sg/prost/services/products/handlers"
    "github.com/sanketh-sg/prost/services/products/repository"
    "github.com/sanketh-sg/prost/services/products/middleware"
    "github.com/sanketh-sg/prost/shared/db"
    "github.com/sanketh-sg/prost/shared/messaging"
)

func main()  {
	//Load env variables
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == ""{
		serviceName = "products"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbSchema := os.Getenv("DB_SCHEMA")
	if dbSchema == "" {
		dbSchema = "catalog"
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
    if rabbitmqURL == "" {
        rabbitmqURL = "amqp://guest:guest@localhost:5672/"
    }

	// Set Gin Mode
	gin.SetMode(gin.ReleaseMode)

	log.Println("=== Products Service Starting ===")
    log.Printf("Service: %s", serviceName)
    log.Printf("Port: %s", port)
    log.Printf("Schema: %s", dbSchema)

	// DB Connection
	log.Println("\nConnecting to PostgreSQL...")
	dbConn, err := db.NewDBConnection(db.Config{
		Host:     "postgres",
        Port:     "5432",
        User:     "prost_admin",
        Password: "prost_password",
        DBName:   "prost",
        Schema:   dbSchema,
	})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbConn.DBConnClose()
	log.Println("Database connected")


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
        // TODO: Implement event subscriber for inventory updates
        // For now, just log that it would listen
        log.Println("✓ Event subscriber ready (TODO: implement handlers)")
    }()

    // Start server in goroutine
    log.Printf("\n✓ Products service listening on :%s", port)
    log.Println("\n=== Service Ready ===\n")

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