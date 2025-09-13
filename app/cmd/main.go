package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"vancouver-trip-planner/internal/handler"
	"vancouver-trip-planner/internal/repository"
	"vancouver-trip-planner/internal/service"
	"vancouver-trip-planner/pkg/maps"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get configuration from environment variables
	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize services
	parkingRepo := repository.NewVancouverParkingRepository()
	pricingService := service.NewPricingService()

	mapsService, err := maps.NewGoogleMapsService(googleMapsAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Google Maps service: %v", err)
	}

	routingService := service.NewRoutingService(parkingRepo, mapsService, pricingService)

	// Initialize handlers
	tripHandler := handler.NewTripHandler(routingService)

	// Setup Gin router
	router := setupRouter(tripHandler)

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(tripHandler *handler.TripHandler) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestIDMiddleware())

	// Health check endpoint
	router.GET("/health", tripHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		trips := v1.Group("/trips")
		{
			trips.POST("/plan", tripHandler.PlanTrip)
		}

		parking := v1.Group("/parking")
		{
			parking.GET("/info", tripHandler.GetParkingInfo)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
