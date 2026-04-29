package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Database
	database.GetDB()
	defer database.CloseDB()

	// Setup Router
	r := gin.Default()

	// CORS Middleware (Basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
	}

	// Protected Routes
	api := r.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/users/me", handlers.GetMe)
		api.GET("/bathrooms/nearby", handlers.GetNearbyBathrooms)
		api.GET("/health/entries", handlers.GetHealthEntries)
	}

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
